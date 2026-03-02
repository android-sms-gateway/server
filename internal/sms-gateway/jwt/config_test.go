package jwt_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     jwt.Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg: jwt.Config{
				Secret:     "01234567890123456789012345678901",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 24 * time.Hour,
				Issuer:     "sms-gate.app",
			},
		},
		{
			name: "refresh not greater than access",
			cfg: jwt.Config{
				Secret:     "01234567890123456789012345678901",
				AccessTTL:  24 * time.Hour,
				RefreshTTL: 24 * time.Hour,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
