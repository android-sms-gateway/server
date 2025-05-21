package settings

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServiceParams struct {
	fx.In

	Repository *repository

	Logger *zap.Logger

	PushSvc *push.Service
}

type Service struct {
	settings *repository

	logger *zap.Logger

	pushSvc *push.Service
}

func (s *Service) GetSettings(userID string, public bool) (map[string]any, error) {
	settings, err := s.settings.GetSettings(userID)
	if err != nil {
		return nil, err
	}

	if !public {
		return settings.Settings, nil
	}

	return filterMap(settings.Settings, rulesPublic)
}

func (s *Service) UpdateSettings(userID string, settings map[string]any) (map[string]any, error) {
	filtered, err := filterMap(settings, rules)
	if err != nil {
		return nil, err
	}

	updatedSettings, err := s.settings.UpdateSettings(&DeviceSettings{
		UserID:   userID,
		Settings: filtered,
	})
	if err != nil {
		return nil, err
	}

	s.notifyDevices(userID)

	return filterMap(updatedSettings.Settings, rulesPublic)
}

func (s *Service) ReplaceSettings(userID string, settings map[string]any) (map[string]any, error) {
	filtered, err := filterMap(settings, rules)
	if err != nil {
		return nil, err
	}

	updated, err := s.settings.ReplaceSettings(&DeviceSettings{
		UserID:   userID,
		Settings: filtered,
	})
	if err != nil {
		return nil, err
	}

	s.notifyDevices(userID)

	return filterMap(updated.Settings, rulesPublic)
}

// notifyDevices asynchronously notifies all the user's devices.
func (s *Service) notifyDevices(userID string) {
	go func(userID string) {
		if err := s.pushSvc.Notify(userID, nil, push.NewSettingsUpdatedEvent()); err != nil {
			s.logger.Error("can't notify devices", zap.Error(err))
		}
	}(userID)
}

func NewService(params ServiceParams) *Service {
	return &Service{
		settings: params.Repository,
		logger:   params.Logger.Named("service"),
		pushSvc:  params.PushSvc,
	}
}
