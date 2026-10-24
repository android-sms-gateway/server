package config_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/server/internal/config"
	infraconfig "github.com/capcom6/go-infra-fx/config"
)

func TestLoadConfigCacheSweepInterval(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		cfg := config.Default()

		if got := time.Duration(cfg.Cache.SweepInterval); got != time.Minute {
			t.Fatalf("expected default sweep interval 1m, got %s", got)
		}
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("CACHE__SWEEP_INTERVAL", "30s")

		cfg := config.Default()
		if err := infraconfig.LoadConfig(&cfg); err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if got := time.Duration(cfg.Cache.SweepInterval); got != 30*time.Second {
			t.Fatalf("expected sweep interval 30s, got %s", got)
		}
	})
}
