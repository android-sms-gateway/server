package events

import (
	"context"
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	DeliveryTypePush = "push"
	DeliveryTypeSSE  = "sse"

	FailureReasonQueueFull      = "queue_full"
	FailureReasonProviderFailed = "provider_failed"
)

type Service struct {
	deviceSvc *devices.Service

	sseSvc  *sse.Service
	pushSvc *push.Service

	queue chan eventWrapper

	enqueuedCounter *prometheus.CounterVec
	sentCounter     *prometheus.CounterVec
	failedCounter   *prometheus.CounterVec

	logger *zap.Logger
}

func NewService(devicesSvc *devices.Service, sseSvc *sse.Service, pushSvc *push.Service, logger *zap.Logger) *Service {
	enqueuedCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "events",
		Name:      "enqueued_total",
		Help:      "Total number of events enqueued",
	}, []string{"event"})

	sentCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "events",
		Name:      "sent_total",
		Help:      "Total number of events sent",
	}, []string{"event", "delivery_type"})

	failedCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "events",
		Name:      "failed_total",
		Help:      "Total number of failed notifications",
	}, []string{"event", "delivery_type", "reason"})

	return &Service{
		deviceSvc: devicesSvc,
		sseSvc:    sseSvc,
		pushSvc:   pushSvc,

		queue: make(chan eventWrapper, 128),

		enqueuedCounter: enqueuedCounter,
		sentCounter:     sentCounter,
		failedCounter:   failedCounter,

		logger: logger,
	}
}

func (s *Service) Notify(userID string, deviceID *string, event *Event) error {
	wrapper := eventWrapper{
		UserID:   userID,
		DeviceID: deviceID,
		Event:    event,
	}

	select {
	case s.queue <- wrapper:
		// Successfully enqueued
		s.enqueuedCounter.WithLabelValues(string(event.eventType)).Inc()
	default:
		s.failedCounter.WithLabelValues(string(event.eventType), FailureReasonQueueFull).Inc()
		return fmt.Errorf("event queue is full")
	}

	return nil
}

func (s *Service) Run(ctx context.Context) {
	for {
		select {
		case wrapper := <-s.queue:
			s.processEvent(wrapper)
		case <-ctx.Done():
			s.logger.Info("Event service stopped")
			return
		}
	}
}

func (s *Service) processEvent(wrapper eventWrapper) {
	// Load devices from database
	filters := []devices.SelectFilter{}
	if wrapper.DeviceID != nil {
		filters = append(filters, devices.WithID(*wrapper.DeviceID))
	}

	devices, err := s.deviceSvc.Select(wrapper.UserID, filters...)
	if err != nil {
		s.logger.Error("Failed to select devices", zap.String("user_id", wrapper.UserID), zap.Error(err))
		return
	}

	if len(devices) == 0 {
		s.logger.Info("No devices found for user", zap.String("user_id", wrapper.UserID))
		return
	}

	// Process each device
	for _, device := range devices {
		if device.PushToken != nil && *device.PushToken != "" {
			// Device has push token, use push service
			if err := s.pushSvc.Enqueue(*device.PushToken, push.Event{
				Type: wrapper.Event.eventType,
				Data: wrapper.Event.data,
			}); err != nil {
				s.logger.Error("Failed to enqueue push notification", zap.String("user_id", wrapper.UserID), zap.String("device_id", device.ID), zap.Error(err))
				s.failedCounter.WithLabelValues(string(wrapper.Event.eventType), DeliveryTypePush, FailureReasonProviderFailed).Inc()
			} else {
				s.sentCounter.WithLabelValues(string(wrapper.Event.eventType), DeliveryTypePush).Inc()
			}
			continue
		}

		// No push token, use SSE service
		if err := s.sseSvc.Send(device.ID, sse.Event{
			Type: wrapper.Event.eventType,
			Data: wrapper.Event.data,
		}); err != nil {
			s.logger.Error("Failed to send SSE notification", zap.String("user_id", wrapper.UserID), zap.String("device_id", device.ID), zap.Error(err))
			s.failedCounter.WithLabelValues(string(wrapper.Event.eventType), DeliveryTypeSSE, FailureReasonProviderFailed).Inc()
		} else {
			s.sentCounter.WithLabelValues(string(wrapper.Event.eventType), DeliveryTypeSSE).Inc()
		}
	}
}
