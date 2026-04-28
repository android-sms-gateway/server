package smpp

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// WebhookMetrics tracks webhook delivery statistics
type WebhookMetrics struct {
	successCount   uint64
	failureCount   uint64
	retryCount     uint64
	lastSuccessAt  int64 // Unix timestamp
	lastFailureAt  int64 // Unix timestamp
	avgLatencyMs   int64 // Average latency in milliseconds
	totalLatencyMs int64 // For calculating average
	latencyCount   uint64
}

// WebhookHandler handles incoming webhooks from the SMS Gateway
type WebhookHandler struct {
	logger  *zap.Logger
	server  *Server
	metrics WebhookMetrics
	mu      sync.RWMutex // Protects metrics
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(logger *zap.Logger, server *Server) *WebhookHandler {
	return &WebhookHandler{
		logger: logger,
		server: server,
	}
}

// Register registers webhook routes on the given router
func (h *WebhookHandler) Register(router fiber.Router) {
	router.Post("/api/smpp/v1/webhook", h.handleWebhook)
	router.Get("/api/smpp/v1/webhook/metrics", h.handleMetrics)
}

// handleWebhook processes incoming webhook requests from the SMS Gateway
func (h *WebhookHandler) handleWebhook(c *fiber.Ctx) error {
	startTime := time.Now()

	sessionID := c.Query("session")
	if sessionID == "" {
		h.updateFailure()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing session parameter",
		})
	}

	var payload webhookPayload
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		h.updateFailure()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	h.logger.Debug("Webhook received",
		zap.String("session", sessionID),
		zap.String("event", payload.Event),
		zap.String("message_id", payload.Message.ID),
		zap.String("state", payload.Message.State),
	)

	// Process the webhook with retry logic
	success := h.processWebhookWithRetry(sessionID, &payload)

	latency := time.Since(startTime)
	if success {
		h.updateSuccess(latency.Milliseconds())
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"latency": latency.Milliseconds(),
		})
	}

	h.updateFailure()
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status": "error",
		"error":  "failed to process webhook",
	})
}

// processWebhookWithRetry attempts to send DELIVER_SM with retries
func (h *WebhookHandler) processWebhookWithRetry(sessionID string, payload *webhookPayload) bool {
	const maxRetries = 3
	const retryDelay = 100 * time.Millisecond

	// Only process delivery-related events
	if payload.Event != "sms:delivery" && payload.Event != "sms:received" {
		h.logger.Debug("Skipping non-delivery event",
			zap.String("event", payload.Event),
			zap.String("session", sessionID),
		)
		return true
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		success, err := h.trySendDeliverSM(sessionID, payload)
		if success {
			return true
		}
		lastErr = err

		if attempt < maxRetries {
			h.logger.Warn("Webhook delivery retry",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.String("session", sessionID),
				zap.Error(err),
			)
			h.updateRetry()
			time.Sleep(retryDelay * time.Duration(attempt)) // Exponential backoff
		}
	}

	h.logger.Error("Webhook delivery failed after retries",
		zap.Int("attempts", maxRetries),
		zap.String("session", sessionID),
		zap.String("message_id", payload.Message.ID),
		zap.Error(lastErr),
	)
	return false
}

// trySendDeliverSM attempts to send a DELIVER_SM PDU to the session
func (h *WebhookHandler) trySendDeliverSM(sessionID string, payload *webhookPayload) (bool, error) {
	session := h.server.GetSession(sessionID)
	if session == nil {
		return false, ErrSessionNotFound
	}

	// Check if session is still bound and active
	if !session.IsBound() {
		return false, ErrSessionNotBound
	}

	// Map gateway message state to SMPP message state
	msgState := mapGatewayStateToSMPP(payload.Message.State)

	session.SendDeliverSM(payload.Message.ID, msgState)
	return true, nil
}

// updateSuccess updates metrics for a successful webhook delivery
func (h *WebhookHandler) updateSuccess(latencyMs int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddUint64(&h.metrics.successCount, 1)
	atomic.StoreInt64(&h.metrics.lastSuccessAt, time.Now().Unix())

	// Update average latency
	oldTotal := atomic.LoadInt64(&h.metrics.totalLatencyMs)
	atomic.StoreInt64(&h.metrics.totalLatencyMs, oldTotal+latencyMs)
	atomic.AddUint64(&h.metrics.latencyCount, 1)
}

// updateFailure updates metrics for a failed webhook delivery
func (h *WebhookHandler) updateFailure() {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddUint64(&h.metrics.failureCount, 1)
	atomic.StoreInt64(&h.metrics.lastFailureAt, time.Now().Unix())
}

// updateRetry updates metrics for a retry attempt
func (h *WebhookHandler) updateRetry() {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddUint64(&h.metrics.retryCount, 1)
}

// GetMetrics returns a copy of the current metrics
func (h *WebhookHandler) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var avgLatency int64
	count := atomic.LoadUint64(&h.metrics.latencyCount)
	if count > 0 {
		avgLatency = atomic.LoadInt64(&h.metrics.totalLatencyMs) / int64(count)
	}

	return map[string]interface{}{
		"success_count":   atomic.LoadUint64(&h.metrics.successCount),
		"failure_count":   atomic.LoadUint64(&h.metrics.failureCount),
		"retry_count":     atomic.LoadUint64(&h.metrics.retryCount),
		"avg_latency_ms":  avgLatency,
		"last_success_at": time.Unix(atomic.LoadInt64(&h.metrics.lastSuccessAt), 0).Format(time.RFC3339),
		"last_failure_at": time.Unix(atomic.LoadInt64(&h.metrics.lastFailureAt), 0).Format(time.RFC3339),
	}
}

// handleMetrics returns webhook delivery metrics
func (h *WebhookHandler) handleMetrics(c *fiber.Ctx) error {
	return c.JSON(h.GetMetrics())
}

// mapGatewayStateToSMPP maps gateway message state string to SMPP message_state
func mapGatewayStateToSMPP(state string) uint8 {
	switch state {
	case "pending", "processed", "sent":
		return 1 // ENROUTE
	case "delivered":
		return 2 // DELIVERED
	case "failed", "rejected":
		return 5 // REJECTED
	case "expired":
		return 4 // EXPIRED
	default:
		return 0 // SCHEDULED
	}
}

// Webhook-specific errors
var (
	ErrSessionNotFound = webhookError("session not found")
	ErrSessionNotBound = webhookError("session not bound")
)

type webhookError string

func (e webhookError) Error() string {
	return string(e)
}

// webhookPayload represents the incoming webhook payload from SMS Gateway
type webhookPayload struct {
	Event   string         `json:"event"`
	Message webhookMessage `json:"message"`
}

// webhookMessage represents the message data in the webhook payload
type webhookMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	DeviceID  string `json:"device_id"`
	SimNumber *int   `json:"sim_number"`
	State     string `json:"state"`
}
