package smpp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"go.uber.org/zap"
)

type Handler struct {
	logger         *zap.Logger
	apiBaseURL     string
	webhookBaseURL string
	httpClient     *http.Client
	authManager    *AuthManager
	server         *Server
	webhook        *WebhookHandler
}

func NewHandler(logger *zap.Logger, apiBaseURL, webhookBaseURL string) *Handler {
	return &Handler{
		logger:         logger,
		apiBaseURL:     apiBaseURL,
		webhookBaseURL: webhookBaseURL,
		httpClient:     &http.Client{},
		authManager:    NewAuthManager(logger, apiBaseURL),
	}
}

// SetServer binds the Server and WebhookHandler to the Handler for metrics aggregation
func (h *Handler) SetServer(server *Server, webhook *WebhookHandler) {
	h.server = server
	h.webhook = webhook
}

func (h *Handler) Authenticate(username, password string) (string, error) {
	// Use AuthManager for token management with refresh support
	tokenInfo, err := h.authManager.Authenticate(username, password)
	if err != nil {
		return "", err
	}

	h.logger.Debug("Token received", zap.String("token", tokenInfo.Token[:20]+"..."))
	return tokenInfo.Token, nil
}

// GetToken returns a valid token, refreshing if necessary
func (h *Handler) GetToken(username, password string) (string, error) {
	return h.authManager.GetToken(username, password)
}

// GetAuthMetrics returns authentication statistics
func (h *Handler) GetAuthMetrics() map[string]interface{} {
	return h.authManager.GetMetrics()
}

// GetMetrics returns combined metrics from all components
func (h *Handler) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"auth": h.authManager.GetMetrics(),
	}

	if h.server != nil {
		metrics["server"] = h.server.GetMetrics()
	}

	if h.webhook != nil {
		metrics["webhook"] = h.webhook.GetMetrics()
	}

	return metrics
}

// RevokeAuthToken removes a cached token
func (h *Handler) RevokeAuthToken(username string) {
	h.authManager.RevokeToken(username)
}

func (h *Handler) SubmitSMS(token string, req *SubmitRequest) (*SubmitResponse, error) {
	client := h.newClient(token)

	msg := smsgateway.Message{
		PhoneNumbers: []string{req.Destination},
		TextMessage: &smsgateway.TextMessage{
			Text: req.Content,
		},
	}

	_, err := client.Send(context.Background(), msg)
	if err != nil {
		// Map error to HTTPError for proper SMPP error mapping
		return nil, NewHTTPError(http.StatusInternalServerError, "submit failed")
	}

	return &SubmitResponse{
		MessageID: "msg-" + req.Destination[:8],
	}, nil
}

func (h *Handler) QuerySMS(token, messageID string) (*QueryResponse, error) {
	client := h.newClient(token)

	msg, err := client.GetState(context.Background(), messageID)
	if err != nil {
		// Map error to HTTPError for proper SMPP error mapping
		return nil, NewHTTPError(http.StatusNotFound, "query failed")
	}

	state := "DELIVERED"
	switch msg.State {
	case smsgateway.ProcessingStatePending, smsgateway.ProcessingStateProcessed, smsgateway.ProcessingStateSent:
		state = string(msg.State)
	case smsgateway.ProcessingStateDelivered:
		state = "DELIVERED"
	case smsgateway.ProcessingStateFailed:
		state = "REJECTED"
	}

	return &QueryResponse{
		MessageID: messageID,
		State:     state,
	}, nil
}

func (h *Handler) RegisterWebhook(token, sessionID string) (string, error) {
	client := h.newClient(token)

	webhook := smsgateway.Webhook{
		URL:   fmt.Sprintf("%s/api/smpp/v1/webhook?session=%s", h.webhookBaseURL, sessionID),
		Event: "sms:received",
	}

	resp, err := client.RegisterWebhook(context.Background(), webhook)
	if err != nil {
		// Map error to HTTPError for proper SMPP error mapping
		return "", NewHTTPError(http.StatusInternalServerError, "register webhook failed")
	}

	return resp.ID, nil
}

func (h *Handler) DeregisterWebhook(token, webhookID string) error {
	if webhookID == "" {
		return nil
	}

	client := h.newClient(token)

	return client.DeleteWebhook(context.Background(), webhookID)
}

func (h *Handler) newClient(token string) *smsgateway.Client {
	return smsgateway.NewClient(smsgateway.Config{
		BaseURL: h.apiBaseURL,
		Token:   token,
	})
}

func (h *Handler) handleWebhookDelivery(payload []byte, sessionID string) error {
	var data WebhookPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	h.logger.Debug("Delivery received",
		zap.String("session", sessionID),
		zap.String("event", data.Event),
		zap.String("message_id", data.Message.ID),
		zap.String("state", data.Message.State),
	)

	return nil
}

type SubmitRequest struct {
	Source      string
	Destination string
	Content     string
}

type SubmitResponse struct {
	MessageID string
}

type QueryResponse struct {
	MessageID string
	State     string
}

type WebhookPayload struct {
	Event   string         `json:"event"`
	Message WebhookMessage `json:"message"`
}

type WebhookMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	DeviceID  string `json:"device_id"`
	SimNumber *int   `json:"sim_number"`
	State     string `json:"state"`
}
