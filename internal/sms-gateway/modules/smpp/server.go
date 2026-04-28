package smpp

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ServerMetrics tracks server-wide statistics
type ServerMetrics struct {
	totalConnections uint64
	activeSessions   uint64
	totalBinds       uint64
	totalUnbinds     uint64
	totalSubmitSM    uint64
	totalDeliverSM   uint64
	totalQuerySM     uint64
	lastConnectionAt int64 // Unix timestamp
	lastBindAt       int64
	lastSubmitAt     int64
}

type Server struct {
	logger      *zap.Logger
	config      Config
	listener    net.Listener
	tlsListener net.Listener
	sessions    map[string]*Session
	sessionsMu  sync.RWMutex
	metrics     ServerMetrics
	quit        chan struct{}
	wg          sync.WaitGroup
}

func NewServer(logger *zap.Logger, config Config) *Server {
	return &Server{
		logger:   logger,
		config:   config,
		sessions: make(map[string]*Session),
		quit:     make(chan struct{}),
	}
}

func (s *Server) Start(ctx context.Context) error {
	if s.config.BindAddress != "" {
		ln, err := net.Listen("tcp", s.config.BindAddress)
		if err != nil {
			return err
		}
		s.listener = ln
		s.logger.Info("SMPP server listening", zap.String("address", s.config.BindAddress))

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.acceptConnections(ctx, ln, false)
		}()
	}

	if s.config.TLSBindAddress != "" && s.config.TLSCert != "" && s.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		ln, err := tls.Listen("tcp", s.config.TLSBindAddress, tlsConfig)
		if err != nil {
			return err
		}
		s.tlsListener = ln
		s.logger.Info("SMPP TLS server listening", zap.String("address", s.config.TLSBindAddress))

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.acceptConnections(ctx, ln, true)
		}()
	}

	return nil
}

// GetServerMetrics returns server-wide metrics
func (s *Server) GetServerMetrics() map[string]interface{} {
	return s.GetMetrics()
}

func (s *Server) Stop(ctx context.Context) error {
	close(s.quit)

	if s.listener != nil {
		_ = s.listener.Close()
	}
	if s.tlsListener != nil {
		_ = s.tlsListener.Close()
	}

	s.wg.Wait()
	return nil
}

// GetSession returns a session by ID, or nil if not found
func (s *Server) GetSession(sessionID string) *Session {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	return s.sessions[sessionID]
}

func (s *Server) acceptConnections(ctx context.Context, ln net.Listener, tls bool) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quit:
			return
		default:
		}

		conn, err := ln.Accept()
		if err != nil {
			if isTemporaryError(err) {
				continue
			}
			return
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(ctx, conn, tls)
		}()
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn, tls bool) {
	defer conn.Close()

	atomic.AddUint64(&s.metrics.totalConnections, 1)
	atomic.StoreInt64(&s.metrics.lastConnectionAt, time.Now().Unix())

	session := NewSession(s.logger, conn, tls, s.config)
	sessionID := session.ID()

	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	atomic.StoreUint64(&s.metrics.activeSessions, uint64(len(s.sessions)))

	defer func() {
		s.sessionsMu.Lock()
		delete(s.sessions, sessionID)
		s.sessionsMu.Unlock()
		atomic.StoreUint64(&s.metrics.activeSessions, uint64(len(s.sessions)))
	}()

	session.Run(ctx)
}

// GetMetrics returns a snapshot of server metrics
func (s *Server) GetMetrics() map[string]interface{} {
	s.sessionsMu.RLock()
	sessionCount := len(s.sessions)
	s.sessionsMu.RUnlock()

	return map[string]interface{}{
		"total_connections": atomic.LoadUint64(&s.metrics.totalConnections),
		"active_sessions":   uint64(sessionCount),
		"total_binds":       atomic.LoadUint64(&s.metrics.totalBinds),
		"total_unbinds":     atomic.LoadUint64(&s.metrics.totalUnbinds),
		"total_submit_sm":   atomic.LoadUint64(&s.metrics.totalSubmitSM),
		"total_deliver_sm":  atomic.LoadUint64(&s.metrics.totalDeliverSM),
		"total_query_sm":    atomic.LoadUint64(&s.metrics.totalQuerySM),
		"last_connection":   time.Unix(atomic.LoadInt64(&s.metrics.lastConnectionAt), 0).Format(time.RFC3339),
	}
}

// IncrementBinds increments the bind counter
func (s *Server) IncrementBinds() {
	atomic.AddUint64(&s.metrics.totalBinds, 1)
	atomic.StoreInt64(&s.metrics.lastBindAt, time.Now().Unix())
}

// IncrementSubmitSM increments the submit counter
func (s *Server) IncrementSubmitSM() {
	atomic.AddUint64(&s.metrics.totalSubmitSM, 1)
	atomic.StoreInt64(&s.metrics.lastSubmitAt, time.Now().Unix())
}

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	ne, ok := err.(net.Error)
	if ok && ne.Temporary() {
		return true
	}
	return strings.Contains(err.Error(), "temporary")
}
