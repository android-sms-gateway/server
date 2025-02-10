package webhooks

import (
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/db"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/capcom6/go-helpers/slices"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServiceParams struct {
	fx.In

	IDGen db.IDGen

	Webhooks *Repository

	DevicesSvc *devices.Service
	PushSvc    *push.Service

	Logger *zap.Logger
}

type Service struct {
	idgen db.IDGen

	webhooks *Repository

	devicesSvc *devices.Service
	pushSvc    *push.Service

	logger *zap.Logger
}

func NewService(params ServiceParams) *Service {
	return &Service{
		idgen:      params.IDGen,
		webhooks:   params.Webhooks,
		devicesSvc: params.DevicesSvc,
		pushSvc:    params.PushSvc,
		logger:     params.Logger,
	}
}

// _select retrieves a list of webhooks that match the provided filters.
func (s *Service) _select(filters ...SelectFilter) ([]smsgateway.Webhook, error) {
	items, err := s.webhooks.Select(filters...)
	if err != nil {
		return nil, fmt.Errorf("can't select webhooks: %w", err)
	}

	return slices.Map(items, webhookToDTO), nil
}

// Select returns a list of webhooks for a specific user that match the provided filters.
// It ensures that the filter includes the user's ID.
func (s *Service) Select(userID string, filters ...SelectFilter) ([]smsgateway.Webhook, error) {
	filters = append(filters, WithUserID(userID))

	return s._select(filters...)
}

// Replace creates or updates a webhook for a given user. After replacing the webhook,
// it asynchronously notifies all the user's devices. Returns an error if the operation fails.
func (s *Service) Replace(userID string, webhook smsgateway.Webhook) error {
	if !smsgateway.IsValidWebhookEvent(webhook.Event) {
		return newValidationError("event", string(webhook.Event), fmt.Errorf("enum value expected"))
	}

	if webhook.ID == "" {
		webhook.ID = s.idgen()
	}

	model := Webhook{
		ExtID:  webhook.ID,
		UserID: userID,
		URL:    webhook.URL,
		Event:  webhook.Event,
	}

	if err := s.webhooks.Replace(&model); err != nil {
		return fmt.Errorf("can't replace webhook: %w", err)
	}

	go s.notifyDevices(userID)

	return nil
}

// Delete removes webhooks for a specific user that match the provided filters.
// It ensures that the filter includes the user's ID.
func (s *Service) Delete(userID string, filters ...SelectFilter) error {
	filters = append(filters, WithUserID(userID))
	if err := s.webhooks.Delete(filters...); err != nil {
		return fmt.Errorf("can't delete webhooks: %w", err)
	}

	go s.notifyDevices(userID)

	return nil
}

// notifyDevices sends a push notification to all devices associated with the given user.
func (s *Service) notifyDevices(userID string) {
	s.logger.Info("Notifying devices", zap.String("user_id", userID))

	devices, err := s.devicesSvc.Select(userID)
	if err != nil {
		s.logger.Error("Failed to select devices", zap.String("user_id", userID), zap.Error(err))
		return
	}

	for _, device := range devices {
		if device.PushToken == nil {
			continue
		}

		if err := s.pushSvc.Enqueue(*device.PushToken, push.NewWebhooksUpdatedEvent()); err != nil {
			s.logger.Error("Failed to send push notification", zap.String("user_id", userID), zap.Error(err))
		}
	}

	s.logger.Info("Notified devices", zap.String("user_id", userID), zap.Int("count", len(devices)))
}
