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
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

const (
	ErrorTTLExpired = "TTL expired"
)

type EnqueueOptions struct {
	SkipPhoneValidation bool
}

type Service struct {
	config Config

	metrics     *metrics
	cache       *cache
	messages    *repository
	hashingTask *hashingTask

	eventsSvc *events.Service

	logger *zap.Logger
	idgen  func() string
}

func NewService(
	config Config,
	metrics *metrics,
	cache *cache,
	messages *repository,
	eventsSvc *events.Service,
	hashingTask *hashingTask,
	logger *zap.Logger,
	idgen db.IDGen,
) *Service {
	return &Service{
		config: config,

		metrics:     metrics,
		cache:       cache,
		messages:    messages,
		hashingTask: hashingTask,

		eventsSvc: eventsSvc,

		logger: logger,
		idgen:  idgen,
	}
}

func (s *Service) RunBackgroundTasks(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.hashingTask.Run(ctx)
	}()
}

func (s *Service) SelectPending(deviceID string, order MessagesOrder) ([]MessageOut, error) {
	if order == "" {
		order = MessagesOrderLIFO
	}

	messages, err := s.messages.SelectPending(deviceID, order)
	if err != nil {
		return nil, err
	}

	return slices.MapOrError(messages, messageToDomain)
}

func (s *Service) UpdateState(device *models.Device, message MessageStateIn) error {
	existing, err := s.messages.Get(MessagesSelectFilter{ExtID: message.ID, DeviceID: device.ID}, MessagesSelectOptions{})
	if err != nil {
		return err
	}

	if message.State == ProcessingStatePending {
		message.State = ProcessingStateProcessed
	}

	existing.State = message.State
	existing.States = slices.Map(maps.Keys(message.States), func(key string) MessageState {
		return MessageState{
			MessageID: existing.ID,
			State:     ProcessingState(key),
			UpdatedAt: message.States[key],
		}
	})
	existing.Recipients = s.recipientsStateToModel(message.Recipients, existing.IsHashed)

	if err := s.messages.UpdateState(&existing); err != nil {
		return err
	}

	if err := s.cache.Set(context.Background(), device.UserID, existing.ExtID, anys.AsPointer(modelToMessageState(existing))); err != nil {
		s.logger.Warn("can't cache message", zap.String("id", existing.ExtID), zap.Error(err))
	}
	s.hashingTask.Enqueue(existing.ID)
	s.metrics.IncTotal(string(existing.State))

	return nil
}

func (s *Service) SelectStates(user models.User, filter MessagesSelectFilter, options MessagesSelectOptions) ([]MessageStateOut, int64, error) {
	filter.UserID = user.ID

	messages, total, err := s.messages.Select(filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("can't select messages: %w", err)
	}

	return slices.Map(messages, modelToMessageState), total, nil
}

func (s *Service) GetState(user models.User, ID string) (*MessageStateOut, error) {
	dto, err := s.cache.Get(context.Background(), user.ID, ID)
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
		MessagesSelectFilter{ExtID: ID, UserID: user.ID},
		MessagesSelectOptions{WithRecipients: true, WithDevice: true, WithStates: true},
	)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			if err := s.cache.Set(context.Background(), user.ID, ID, nil); err != nil {
				s.logger.Warn("can't cache message", zap.String("id", ID), zap.Error(err))
			}
		}

		return nil, err
	}

	dto = anys.AsPointer(modelToMessageState(message))
	if err := s.cache.Set(context.Background(), user.ID, ID, dto); err != nil {
		s.logger.Warn("can't cache message", zap.String("id", ID), zap.Error(err))
	}

	return dto, nil
}

func (s *Service) Enqueue(device models.Device, message MessageIn, opts EnqueueOptions) (MessageStateOut, error) {
	state := MessageStateOut{
		DeviceID: device.ID,
		MessageStateIn: MessageStateIn{
			State:      ProcessingStatePending,
			Recipients: make([]smsgateway.RecipientState, len(message.PhoneNumbers)),
		},
	}

	var phone string
	var err error
	for i, v := range message.PhoneNumbers {
		if message.IsEncrypted || opts.SkipPhoneValidation {
			phone = v
		} else {
			if phone, err = cleanPhoneNumber(v); err != nil {
				return state, fmt.Errorf("can't use phone in row %d: %w", i+1, err)
			}
		}

		message.PhoneNumbers[i] = phone

		state.Recipients[i] = smsgateway.RecipientState{
			PhoneNumber: phone,
			State:       smsgateway.ProcessingStatePending,
		}
	}

	validUntil := message.ValidUntil
	if message.TTL != nil && *message.TTL > 0 {
		validUntil = anys.AsPointer(time.Now().Add(time.Duration(*message.TTL) * time.Second))
	}

	msg := Message{
		ExtID:       message.ID,
		Recipients:  s.recipientsToModel(message.PhoneNumbers),
		IsEncrypted: message.IsEncrypted,

		DeviceID: device.ID,

		SimNumber:          message.SimNumber,
		WithDeliveryReport: anys.OrDefault(message.WithDeliveryReport, true),

		Priority:   int8(message.Priority),
		ValidUntil: validUntil,
	}

	if message.TextContent != nil {
		if err := msg.SetTextContent(*message.TextContent); err != nil {
			return state, fmt.Errorf("can't set text content: %w", err)
		}
	} else if message.DataContent != nil {
		if err := msg.SetDataContent(*message.DataContent); err != nil {
			return state, fmt.Errorf("can't set data content: %w", err)
		}
	} else {
		return state, errors.New("no text or data content")
	}

	if msg.ExtID == "" {
		msg.ExtID = s.idgen()
	}
	state.ID = msg.ExtID

	if err := s.messages.Insert(&msg); err != nil {
		return state, err
	}

	if err := s.cache.Set(context.Background(), device.UserID, message.ID, anys.AsPointer(modelToMessageState(msg))); err != nil {
		s.logger.Warn("can't cache message", zap.String("id", msg.ExtID), zap.Error(err))
	}
	s.metrics.IncTotal(string(msg.State))

	go func(userID, deviceID string) {
		if err := s.eventsSvc.Notify(userID, &deviceID, events.NewMessageEnqueuedEvent()); err != nil {
			s.logger.Error("can't notify device", zap.Error(err), zap.String("user_id", userID), zap.String("device_id", deviceID))
		}
	}(device.UserID, device.ID)

	return state, nil
}

func (s *Service) ExportInbox(device models.Device, since, until time.Time) error {
	event := events.NewMessagesExportRequestedEvent(since, until)

	return s.eventsSvc.Notify(device.UserID, &device.ID, event)
}

func (s *Service) Clean(ctx context.Context) error {
	//TODO: use delete queue to optimize deletion
	n, err := s.messages.removeProcessed(ctx, time.Now().Add(-s.config.ProcessedLifetime))

	s.logger.Info("Cleaned processed messages", zap.Int64("count", n))
	return err
}

///////////////////////////////////////////////////////////////////////////////

func (s *Service) recipientsToModel(input []string) []MessageRecipient {
	output := make([]MessageRecipient, len(input))

	for i, v := range input {
		output[i] = MessageRecipient{
			PhoneNumber: v,
		}
	}

	return output
}

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

		output[i] = MessageRecipient{
			PhoneNumber: phoneNumber,
			State:       ProcessingState(v.State),
			Error:       v.Error,
		}
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
		return input, ErrValidation(fmt.Sprintf("can't parse phone number: %s", err.Error()))
	}

	if !phonenumbers.IsValidNumber(phone) {
		return input, ErrValidation("invalid phone number")
	}

	phoneNumberType := phonenumbers.GetNumberType(phone)
	if phoneNumberType != phonenumbers.MOBILE && phoneNumberType != phonenumbers.FIXED_LINE_OR_MOBILE {
		return input, ErrValidation("not mobile phone number")
	}

	return phonenumbers.Format(phone, phonenumbers.E164), nil
}
