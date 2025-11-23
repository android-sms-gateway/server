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
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
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

	metrics       *metrics
	cache         *cache
	messages      *Repository
	hashingWorker *hashingWorker

	eventsSvc *events.Service

	logger *zap.Logger
	idgen  func() string
}

func NewService(
	config Config,
	metrics *metrics,
	cache *cache,
	messages *Repository,
	eventsSvc *events.Service,
	hashingTask *hashingWorker,
	logger *zap.Logger,
	idgen db.IDGen,
) *Service {
	return &Service{
		config: config,

		metrics:       metrics,
		cache:         cache,
		messages:      messages,
		hashingWorker: hashingTask,

		eventsSvc: eventsSvc,

		logger: logger,
		idgen:  idgen,
	}
}

func (s *Service) RunBackgroundTasks(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.hashingWorker.Run(ctx)
	}()
}

func (s *Service) SelectPending(deviceID string, order Order) ([]MessageOut, error) {
	if order == "" {
		order = MessagesOrderLIFO
	}

	messages, err := s.messages.SelectPending(deviceID, order)
	if err != nil {
		return nil, err
	}

	return slices.MapOrError(messages, messageToDomain) //nolint:wrapcheck // already wrapped
}

func (s *Service) UpdateState(device *models.Device, message MessageStateIn) error {
	existing, err := s.messages.Get(
		*new(SelectFilter).WithExtID(message.ID).WithDeviceID(device.ID),
		SelectOptions{}, //nolint:exhaustruct // not needed
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
		func(key string, value time.Time) MessageState {
			return MessageState{
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

	if cacheErr := s.cache.Set(context.Background(), device.UserID, existing.ExtID, anys.AsPointer(modelToMessageState(existing))); cacheErr != nil {
		s.logger.Warn("failed to cache message", zap.String("id", existing.ExtID), zap.Error(cacheErr))
	}
	s.hashingWorker.Enqueue(existing.ID)
	s.metrics.IncTotal(string(existing.State))

	return nil
}

func (s *Service) SelectStates(
	user users.User,
	filter SelectFilter,
	options SelectOptions,
) ([]MessageStateOut, int64, error) {
	filter.UserID = user.ID

	messages, total, err := s.messages.Select(filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select messages: %w", err)
	}

	return slices.Map(messages, modelToMessageState), total, nil
}

func (s *Service) GetState(user users.User, id string) (*MessageStateOut, error) {
	dto, err := s.cache.Get(context.Background(), user.ID, id)
	if err == nil {
		s.metrics.IncCache(true)

		// Cache nil entries represent "not found" and prevent repeated lookups
		if dto == nil {
			return nil, ErrMessageNotFound
		}
		return dto, nil
	}
	s.metrics.IncCache(false)

	message, err := s.messages.Get(
		*new(SelectFilter).WithExtID(id).WithUserID(user.ID),
		*new(SelectOptions).IncludeRecipients().IncludeDevice().IncludeStates(),
	)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			if cacheErr := s.cache.Set(context.Background(), user.ID, id, nil); cacheErr != nil {
				s.logger.Warn("failed to cache message", zap.String("id", id), zap.Error(cacheErr))
			}
		}

		return nil, err
	}

	dto = anys.AsPointer(modelToMessageState(message))
	if cacheErr := s.cache.Set(context.Background(), user.ID, id, dto); cacheErr != nil {
		s.logger.Warn("failed to cache message", zap.String("id", id), zap.Error(cacheErr))
	}

	return dto, nil
}

func (s *Service) Enqueue(device models.Device, message MessageIn, opts EnqueueOptions) (*MessageStateOut, error) {
	msg, err := s.prepareMessage(device, message, opts)
	if err != nil {
		return nil, err
	}

	state := &MessageStateOut{
		DeviceID: device.ID,
		MessageStateIn: MessageStateIn{
			ID:    msg.ExtID,
			State: ProcessingStatePending,
			Recipients: lo.Map(
				msg.Recipients,
				func(item MessageRecipient, _ int) smsgateway.RecipientState { return modelToRecipientState(item) },
			),
			States: map[string]time.Time{},
		},
		IsHashed:    false,
		IsEncrypted: msg.IsEncrypted,
	}

	if insErr := s.messages.Insert(msg); insErr != nil {
		return state, insErr
	}

	if cacheErr := s.cache.Set(context.Background(), device.UserID, msg.ExtID, anys.AsPointer(modelToMessageState(*msg))); cacheErr != nil {
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

func (s *Service) prepareMessage(device models.Device, message MessageIn, opts EnqueueOptions) (*Message, error) {
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

	msg := NewMessage(
		message.ID,
		device.ID,
		message.PhoneNumbers,
		int8(message.Priority),
		message.SimNumber,
		validUntil,
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

func (s *Service) ExportInbox(device models.Device, since, until time.Time) error {
	event := events.NewMessagesExportRequestedEvent(since, until)

	if err := s.eventsSvc.Notify(device.UserID, &device.ID, event); err != nil {
		return fmt.Errorf("failed to notify device: %w", err)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (s *Service) recipientsStateToModel(input []smsgateway.RecipientState, hash bool) []MessageRecipient {
	output := make([]MessageRecipient, len(input))

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

func modelToMessageState(input Message) MessageStateOut {
	return MessageStateOut{
		DeviceID:    input.DeviceID,
		IsHashed:    input.IsHashed,
		IsEncrypted: input.IsEncrypted,

		MessageStateIn: MessageStateIn{
			ID:         input.ExtID,
			State:      input.State,
			Recipients: slices.Map(input.Recipients, modelToRecipientState),
			States: slices.Associate(
				input.States,
				func(state MessageState) string { return string(state.State) },
				func(state MessageState) time.Time { return state.UpdatedAt },
			),
		},
	}
}

func modelToRecipientState(input MessageRecipient) smsgateway.RecipientState {
	return smsgateway.RecipientState{
		PhoneNumber: input.PhoneNumber,
		State:       smsgateway.ProcessingState(input.State),
		Error:       input.Error,
	}
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
