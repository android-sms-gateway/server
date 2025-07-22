package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type Service struct {
	config Config

	mu          sync.RWMutex
	connections map[string]*sseConnection

	logger *zap.Logger
}

type sseConnection struct {
	channel     chan []byte
	closeSignal chan struct{}
}

func NewService(config Config, logger *zap.Logger) *Service {
	return &Service{
		config: config,

		connections: make(map[string]*sseConnection),

		logger: logger,
	}
}

func (s *Service) Send(ctx context.Context, messages map[string]domain.Event) (map[string]error, error) {
	errs := make(map[string]error)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for deviceId, event := range messages {
		conn, exists := s.connections[deviceId]
		if !exists {
			errs[deviceId] = fmt.Errorf("client not connected")
			s.logger.Debug("Client not connected", zap.String("client_id", deviceId))
			continue
		}

		data, err := json.Marshal(event.Map())
		if err != nil {
			errs[deviceId] = fmt.Errorf("can't marshal payload: %w", err)
			s.logger.Error("Failed to marshal event for client",
				zap.String("client_id", deviceId),
				zap.Any("event", event),
				zap.Error(err))
			continue
		}

		select {
		case conn.channel <- data:
			// Message sent successfully
		case <-ctx.Done():
			errs[deviceId] = ctx.Err()
			s.logger.Warn("Failed to send event to client",
				zap.String("client_id", deviceId),
				zap.Error(ctx.Err()))
		case <-conn.closeSignal:
			errs[deviceId] = fmt.Errorf("connection closed")
			s.logger.Warn("Failed to send event to client",
				zap.String("client_id", deviceId),
				zap.Error(fmt.Errorf("connection closed")))
		}
	}

	return errs, nil
}

func (s *Service) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, conn := range s.connections {
		close(conn.closeSignal)
		delete(s.connections, id)
	}
	return nil
}

func (s *Service) Handler(deviceId string, c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		s.registerConnection(deviceId)
		defer s.removeConnection(deviceId)

		conn := s.getConnection(deviceId)
		if conn == nil {
			s.logger.Warn("Client not connected", zap.String("client_id", deviceId))
			return
		}

		if err := s.writeToStream(w, ":keepalive"); err != nil {
			s.logger.Warn("Failed to write keepalive",
				zap.String("client_id", deviceId),
				zap.Error(err))
			return
		}

		ticker := time.NewTicker(s.config.keepAlivePeriod)
		defer ticker.Stop()

		for {
			select {
			case data := <-conn.channel:
				if err := s.writeToStream(w, fmt.Sprintf("data: %s", utils.UnsafeString(data))); err != nil {
					s.logger.Warn("Failed to write event data",
						zap.String("client_id", deviceId),
						zap.Error(err))
					return
				}
			case <-ticker.C:
				if err := s.writeToStream(w, ":keepalive"); err != nil {
					s.logger.Warn("Failed to write keepalive",
						zap.String("client_id", deviceId),
						zap.Error(err))
					return
				}
			case <-conn.closeSignal:
				return
			}
		}
	})

	return nil
}

func (s *Service) writeToStream(w *bufio.Writer, data string) error {
	if _, err := fmt.Fprintf(w, "%s\n\n", data); err != nil {
		return err
	}
	return w.Flush()
}

func (s *Service) registerConnection(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingConn, ok := s.connections[id]; ok {
		s.logger.Info("Closing existing SSE connection", zap.String("client_id", id))
		close(existingConn.closeSignal)
		delete(s.connections, id)
	}

	s.connections[id] = &sseConnection{
		channel:     make(chan []byte, 8),
		closeSignal: make(chan struct{}),
	}
	s.logger.Info("Registering SSE connection", zap.String("client_id", id))
}

func (s *Service) removeConnection(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.connections[id]; exists {
		close(conn.closeSignal)
		delete(s.connections, id)
		s.logger.Info("Removing SSE connection", zap.String("client_id", id))
	}
}

func (s *Service) getConnection(id string) *sseConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connections[id]
}
