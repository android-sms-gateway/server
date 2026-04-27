package messages

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/db"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/events"
	"github.com/capcom6/go-helpers/anys"
	"github.com/capcom6/go-helpers/slices"
	"github.com/nyaruka/phonenumbers"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type EnqueueOptions struct {
	SkipPhoneValidation bool
}

type Service struct {
	config Config

	limiter   *Limiter
	messages  *Repository
	eventsSvc *events.Service

	metrics       *metrics
	cache         *stateCache
	hashingWorker *hashingWorker

	logger *zap.Logger
	idgen  func() string
}

func NewService(
	config Config,

	limiter *Limiter,
	messages *Repository,
	eventsSvc *events.Service,

	metrics *metrics,
	cache *stateCache,
	hashingTask *hashingWorker,

	logger *zap.Logger,
	idgen db.IDGen,
) *Service {
	return &Service{
		config: config,

		limiter:   limiter,
		messages:  messages,
		eventsSvc: eventsSvc,

		metrics:       metrics,
		cache:         cache,
		hashingWorker: hashingTask,

		logger: logger,
		idgen:  idgen,
	}
}

func (s *Service) RunBackgroundTasks(ctx context.Context, wg *sync.WaitGroup) {
	wg.Go(func() {
		s.hashingWorker.Run(ctx)
	})
	wg.Go(func() {
		s.limiter.Run(ctx)
	})
}

func (s *Service) SelectPending(deviceID string, order Order) ([]Message, error) {
	if order == "" {
		order = MessagesOrderLIFO
	}

	messages, err := s.messages.listPending(deviceID, order)
	if err != nil {
		return nil, err
	}

	return slices.MapOrError(messages, messageToDomain) //nolint:wrapcheck // already wrapped
}

func (s *Service) UpdateState(device *models.Device, message MessageStateInput) error {
	existing, err := s.messages.get(
		*new(SelectFilter).WithExtID(message.ID).WithDeviceID(device.ID),
		*new(SelectOptions).IncludeContent(),
	)
	if err != nil {
		return err
	}

	if message.State == ProcessingStatePending {
		message.State = ProcessingStateProcessed
	}

	existing.State = message.State
	existing.States = lo.MapToSlice(
		message.States,
		func(key string, value time.Time) messageStateModel {
			return messageStateModel{
				ID:        0,
				MessageID: existing.ID,
				State:     ProcessingState(key),
				UpdatedAt: value,
			}
		},
	)
	existing.Recipients = s.recipientsStateToModel(message.Recipients, existing.IsHashed)

	if updErr := s.messages.UpdateState(&existing); updErr != nil {
		return updErr
	}

	state, err := existing.toStateDomain()
	if err != nil {
		return err
	}

	if cacheErr := s.cache.Set(
		context.Background(),
		device.UserID,
		existing.ExtID,
		state,
	); cacheErr != nil {
		s.logger.Warn("failed to cache message", zap.String("id", existing.ExtID), zap.Error(cacheErr))
	}
	s.hashingWorker.Enqueue(existing.ID)
	s.metrics.IncTotal(string(existing.State))

	return nil
}

func (s *Service) SelectStates(
	userID string,
	filter SelectFilter,
	options SelectOptions,
) ([]MessageState, int64, error) {
	filter.UserID = userID

	messages, total, err := s.messages.list(filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select messages: %w", err)
	}

	result, err := slices.MapOrError(
		messages,
		func(m messageModel) (*MessageState, error) {
			return m.toStateDomain()
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to map messages: %w", err)
	}

	return lo.FromSlicePtr(result), total, nil
}

func (s *Service) GetState(userID string, id string) (*MessageState, error) {
	state, err := s.cache.Get(context.Background(), userID, id)
	if err == nil {
		s.metrics.IncCache(true)

		// Cache nil entries represent "not found" and prevent repeated lookups
		if state == nil {
			return nil, ErrMessageNotFound
		}
		return state, nil
	}
	s.metrics.IncCache(false)

	message, err := s.messages.get(
		*new(SelectFilter).WithExtID(id).WithUserID(userID),
		*new(SelectOptions).IncludeRecipients().IncludeDevice().IncludeStates().IncludeContent(),
	)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			if cacheErr := s.cache.Set(context.Background(), userID, id, nil); cacheErr != nil {
				s.logger.Warn("failed to cache message", zap.String("id", id), zap.Error(cacheErr))
			}
		}

		return nil, err
	}

	state, err = message.toStateDomain()
	if err != nil {
		return nil, err
	}

	if cacheErr := s.cache.Set(context.Background(), userID, id, state); cacheErr != nil {
		s.logger.Warn("failed to cache message", zap.String("id", id), zap.Error(cacheErr))
	}

	return state, nil
}

func (s *Service) Enqueue(
	ctx context.Context,
	device models.Device,
	message MessageInput,
	opts EnqueueOptions,
) (*MessageState, error) {
	// Trigger async queue stats refresh for this device
	if err := s.limiter.Refresh(ctx, device.ID); err != nil {
		s.logger.Error("failed to refresh queue stats", zap.String("device_id", device.ID), zap.Error(err))
	}

	if err := s.limiter.Check(ctx, device.ID); err != nil {
		return nil, err
	}

	msg, err := s.prepareMessage(device, message, opts)
	if err != nil {
		return nil, err
	}

	state, err := msg.toStateDomain()
	if err != nil {
		return nil, err
	}

	if insErr := s.messages.Insert(msg); insErr != nil {
		return state, insErr
	}

	if cacheErr := s.cache.Set(
		context.Background(),
		device.UserID,
		msg.ExtID,
		state,
	); cacheErr != nil {
		s.logger.Warn("failed to cache message", zap.String("id", msg.ExtID), zap.Error(cacheErr))
	}
	s.metrics.IncTotal(string(msg.State))

	go func(userID, deviceID string) {
		if ntfErr := s.eventsSvc.Notify(userID, &deviceID, events.NewMessageEnqueuedEvent()); ntfErr != nil {
			s.logger.Error(
				"failed to notify device",
				zap.Error(ntfErr),
				zap.String("user_id", userID),
				zap.String("device_id", deviceID),
			)
		}
	}(device.UserID, device.ID)

	return state, nil
}

func (s *Service) prepareMessage(
	device models.Device,
	message MessageInput,
	opts EnqueueOptions,
) (*messageModel, error) {
	var phone string
	var err error
	for i, v := range message.PhoneNumbers {
		if message.IsEncrypted || opts.SkipPhoneValidation {
			phone = v
		} else {
			if phone, err = cleanPhoneNumber(v); err != nil {
				return nil, fmt.Errorf("failed to use phone in row %d: %w", i+1, err)
			}
		}

		message.PhoneNumbers[i] = phone
	}

	validUntil := message.ValidUntil
	if message.TTL != nil && *message.TTL > 0 {
		//nolint:gosec // not a problem
		validUntil = anys.AsPointer(
			time.Now().Add(time.Duration(*message.TTL) * time.Second),
		)
	}

	if message.ScheduleAt != nil && validUntil != nil && message.ScheduleAt.After(*validUntil) {
		return nil, ValidationError("scheduleAt must be less than or equal to validUntil")
	}

	msg := newMessageModel(
		message.ID,
		device.ID,
		message.PhoneNumbers,
		int8(message.Priority),
		message.SimNumber,
		validUntil,
		message.ScheduleAt,
		anys.OrDefault(message.WithDeliveryReport, true),
		message.IsEncrypted,
	)

	switch {
	case message.TextContent != nil:
		if setErr := msg.SetTextContent(*message.TextContent); setErr != nil {
			return nil, fmt.Errorf("failed to set text content: %w", setErr)
		}
	case message.DataContent != nil:
		if setErr := msg.SetDataContent(*message.DataContent); setErr != nil {
			return nil, fmt.Errorf("failed to set data content: %w", setErr)
		}
	default:
		return nil, ErrNoContent
	}

	if msg.ExtID == "" {
		msg.ExtID = s.idgen()
	}

	return msg, nil
}

/////////////////////////////////////////////////////////////////////////////

func (s *Service) recipientsStateToModel(input []smsgateway.RecipientState, hash bool) []messageRecipientModel {
	output := make([]messageRecipientModel, len(input))

	for i, v := range input {
		phoneNumber := v.PhoneNumber
		if len(phoneNumber) > 0 && phoneNumber[0] != '+' {
			// compatibility with Android app before 1.1.1
			phoneNumber = "+" + phoneNumber
		}

		if v.State == smsgateway.ProcessingStatePending {
			v.State = smsgateway.ProcessingStateProcessed
		}

		if hash {
			phoneNumber = fmt.Sprintf("%x", sha256.Sum256([]byte(phoneNumber)))[:16]
		}

		output[i] = newMessageRecipient(
			phoneNumber,
			ProcessingState(v.State),
			v.Error,
		)
	}

	return output
}

func cleanPhoneNumber(input string) (string, error) {
	phone, err := phonenumbers.Parse(input, "RU")
	if err != nil {
		return input, ValidationError(fmt.Sprintf("failed to parse phone number: %s", err.Error()))
	}

	if !phonenumbers.IsValidNumber(phone) {
		return input, ValidationError("invalid phone number")
	}

	phoneNumberType := phonenumbers.GetNumberType(phone)
	if phoneNumberType != phonenumbers.MOBILE && phoneNumberType != phonenumbers.FIXED_LINE_OR_MOBILE {
		return input, ValidationError("not mobile phone number")
	}

	return phonenumbers.Format(phone, phonenumbers.E164), nil
}
