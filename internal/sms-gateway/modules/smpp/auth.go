package smpp

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"go.uber.org/zap"
)

// AuthMetrics tracks authentication statistics
type AuthMetrics struct {
	successCount   uint64
	failureCount   uint64
	refreshCount   uint64
	lastSuccessAt  int64 // Unix timestamp
	lastFailureAt  int64 // Unix timestamp
	avgLatencyMs   int64 // Average latency in milliseconds (atomic)
	totalLatencyMs int64 // For calculating average
	latencyCount   uint64
}

// TokenInfo holds JWT token with expiry information
type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
	Username  string
	mu        sync.RWMutex
}

// IsExpired checks if the token is expired or about to expire within the buffer period
func (t *TokenInfo) IsExpired(buffer time.Duration) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(buffer).After(t.ExpiresAt)
}

// GetToken returns the current token if valid
func (t *TokenInfo) GetToken() (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.Token == "" {
		return "", false
	}
	if t.IsExpired(30 * time.Second) {
		return "", false
	}
	return t.Token, true
}

// Update updates the token and expiry time
func (t *TokenInfo) Update(token string, expiresAt time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Token = token
	t.ExpiresAt = expiresAt
}

// AuthManager handles authentication with token refresh and metrics
type AuthManager struct {
	logger     *zap.Logger
	apiBaseURL string
	metrics    AuthMetrics
	tokens     sync.Map // map[username]*TokenInfo
}

// NewAuthManager creates a new AuthManager
func NewAuthManager(logger *zap.Logger, apiBaseURL string) *AuthManager {
	return &AuthManager{
		logger:     logger,
		apiBaseURL: apiBaseURL,
		metrics:    AuthMetrics{},
	}
}

// Authenticate authenticates a user and returns a token with expiry
func (a *AuthManager) Authenticate(username, password string) (*TokenInfo, error) {
	start := time.Now()

	config := smsgateway.Config{
		BaseURL:  a.apiBaseURL,
		User:     username,
		Password: password,
	}

	client := smsgateway.NewClient(config)
	resp, err := client.GenerateToken(context.Background(), smsgateway.TokenRequest{
		TTL: 3600, // 1 hour
	})

	latency := time.Since(start).Milliseconds()
	a.recordLatency(latency)

	if err != nil {
		atomic.AddUint64(&a.metrics.failureCount, 1)
		atomic.StoreInt64(&a.metrics.lastFailureAt, time.Now().Unix())
		a.logger.Error("Authentication failed",
			zap.Error(err),
			zap.String("username", username),
			zap.Int64("latency_ms", latency),
		)
		return nil, NewHTTPError(http.StatusUnauthorized, "authentication failed")
	}

	// Use the ExpiresAt from response with a 30s buffer
	expiresAt := resp.ExpiresAt.Add(-30 * time.Second)

	tokenInfo := &TokenInfo{
		Token:     resp.AccessToken,
		Username:  username,
		ExpiresAt: expiresAt,
	}

	// Store token
	a.tokens.Store(username, tokenInfo)

	atomic.AddUint64(&a.metrics.successCount, 1)
	atomic.StoreInt64(&a.metrics.lastSuccessAt, time.Now().Unix())

	a.logger.Info("Authentication successful",
		zap.String("username", username),
		zap.Int64("latency_ms", latency),
		zap.Time("expires_at", expiresAt),
	)

	return tokenInfo, nil
}

// GetToken returns a valid token for the user, refreshing if necessary
func (a *AuthManager) GetToken(username, password string) (string, error) {
	// Check if we have a valid cached token
	if existing, ok := a.tokens.Load(username); ok {
		tokenInfo := existing.(*TokenInfo)
		if token, valid := tokenInfo.GetToken(); valid {
			return token, nil
		}
	}

	// Need to authenticate/refresh
	tokenInfo, err := a.Authenticate(username, password)
	if err != nil {
		return "", err
	}

	atomic.AddUint64(&a.metrics.refreshCount, 1)
	return tokenInfo.Token, nil
}

// GetMetrics returns a copy of the current metrics
func (a *AuthManager) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"success_count":   atomic.LoadUint64(&a.metrics.successCount),
		"failure_count":   atomic.LoadUint64(&a.metrics.failureCount),
		"refresh_count":   atomic.LoadUint64(&a.metrics.refreshCount),
		"last_success_at": atomic.LoadInt64(&a.metrics.lastSuccessAt),
		"last_failure_at": atomic.LoadInt64(&a.metrics.lastFailureAt),
		"avg_latency_ms":  atomic.LoadInt64(&a.metrics.avgLatencyMs),
	}
}

// recordLatency updates the average latency
func (a *AuthManager) recordLatency(latencyMs int64) {
	totalLatency := atomic.LoadInt64(&a.metrics.totalLatencyMs) + latencyMs
	count := atomic.AddUint64(&a.metrics.latencyCount, 1)
	atomic.StoreInt64(&a.metrics.totalLatencyMs, totalLatency)
	atomic.StoreInt64(&a.metrics.avgLatencyMs, totalLatency/int64(count))
}

// ValidateCredentials performs a lightweight validation of credentials
// without necessarily creating a full session
func (a *AuthManager) ValidateCredentials(username, password string) bool {
	_, err := a.Authenticate(username, password)
	return err == nil
}

// RevokeToken removes a cached token (called on UNBIND)
func (a *AuthManager) RevokeToken(username string) {
	a.tokens.Delete(username)
}

// TokenCount returns the number of cached tokens
func (a *AuthManager) TokenCount() int {
	count := 0
	a.tokens.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
