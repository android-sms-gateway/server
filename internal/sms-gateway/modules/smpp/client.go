package smpp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"go.uber.org/zap"
)

// Client wraps the SMS Gateway client for SMPP operations
type Client struct {
	logger     *zap.Logger
	apiBaseURL string
	httpClient *http.Client
}

// NewClient creates a new SMS Gateway client wrapper
func NewClient(logger *zap.Logger, apiBaseURL string) *Client {
	return &Client{
		logger:     logger,
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{},
	}
}

// Authenticate validates credentials and returns a JWT token
func (c *Client) Authenticate(username, password string) (string, error) {
	config := smsgateway.Config{
		BaseURL:  c.apiBaseURL,
		User:     username,
		Password: password,
	}

	client := smsgateway.NewClient(config)
	resp, err := client.GenerateToken(context.Background(), smsgateway.TokenRequest{
		TTL: 3600,
	})
	if err != nil {
		c.logger.Error("Authentication failed",
			zap.Error(err),
			zap.String("username", username),
		)
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	c.logger.Info("Authentication successful",
		zap.String("username", username),
	)

	return resp.AccessToken, nil
}

// createClient creates an authenticated SMS Gateway client
func (c *Client) createClient(token string) *smsgateway.Client {
	return smsgateway.NewClient(smsgateway.Config{
		BaseURL: c.apiBaseURL,
		Token:   token,
	})
}

// SubmitMessage sends an SMS message via the gateway
func (c *Client) SubmitMessage(token, destination, content string) (string, error) {
	client := c.createClient(token)

	msg := smsgateway.Message{
		PhoneNumbers: []string{destination},
		TextMessage: &smsgateway.TextMessage{
			Text: content,
		},
	}

	_, err := client.Send(context.Background(), msg)
	if err != nil {
		c.logger.Error("Message submit failed",
			zap.Error(err),
			zap.String("destination", destination),
		)
		return "", fmt.Errorf("submit failed: %w", err)
	}

	// Generate a message ID for the response
	messageID := fmt.Sprintf("msg-%s", destination[:min(8, len(destination))])
	c.logger.Info("Message submitted",
		zap.String("message_id", messageID),
		zap.String("destination", destination),
	)

	return messageID, nil
}

// QueryMessageStatus queries the status of a sent message
func (c *Client) QueryMessageStatus(token, messageID string) (string, error) {
	client := c.createClient(token)

	msg, err := client.GetState(context.Background(), messageID)
	if err != nil {
		c.logger.Error("Message query failed",
			zap.Error(err),
			zap.String("message_id", messageID),
		)
		return "", fmt.Errorf("query failed: %w", err)
	}

	state := mapMessageState(msg.State)
	return state, nil
}

// RegisterWebhook registers a webhook for delivery receipts
func (c *Client) RegisterWebhook(token, url string) (string, error) {
	client := c.createClient(token)

	webhook := smsgateway.Webhook{
		URL:   url,
		Event: "sms:received",
	}

	resp, err := client.RegisterWebhook(context.Background(), webhook)
	if err != nil {
		c.logger.Error("Webhook registration failed",
			zap.Error(err),
			zap.String("url", url),
		)
		return "", fmt.Errorf("webhook registration failed: %w", err)
	}

	c.logger.Info("Webhook registered",
		zap.String("webhook_id", resp.ID),
		zap.String("url", url),
	)

	return resp.ID, nil
}

// DeregisterWebhook removes a previously registered webhook
func (c *Client) DeregisterWebhook(token, webhookID string) error {
	if webhookID == "" {
		return nil
	}

	client := c.createClient(token)

	err := client.DeleteWebhook(context.Background(), webhookID)
	if err != nil {
		c.logger.Error("Webhook deregistration failed",
			zap.Error(err),
			zap.String("webhook_id", webhookID),
		)
		return fmt.Errorf("webhook deregistration failed: %w", err)
	}

	c.logger.Info("Webhook deregistered",
		zap.String("webhook_id", webhookID),
	)

	return nil
}

// mapMessageState converts gateway message state to SMPP message state
func mapMessageState(state smsgateway.ProcessingState) string {
	switch state {
	case smsgateway.ProcessingStatePending:
		return "PENDING"
	case smsgateway.ProcessingStateProcessed:
		return "PROCESSED"
	case smsgateway.ProcessingStateSent:
		return "SENT"
	case smsgateway.ProcessingStateDelivered:
		return "DELIVERED"
	case smsgateway.ProcessingStateFailed:
		return "REJECTED"
	default:
		return "UNKNOWN"
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
