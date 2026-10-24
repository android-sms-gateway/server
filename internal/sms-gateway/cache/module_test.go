package cache_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	appcache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// TestModuleRunsSweeper boots the cache module together with the cachefx
// module the way the app wires them, and checks the prefixed factory resolves
// and the sweeper runnable starts and stops cleanly.
func TestModuleRunsSweeper(t *testing.T) {
	app := fx.New(
		fx.Supply(zap.NewNop()),
		fx.Supply(cachefx.Config{URL: "memory://", SweepInterval: 10 * time.Millisecond}),
		cachefx.Module(cachefx.WithSweeper()),
		appcache.Module(),
		fx.Invoke(func(factory appcache.Factory) error {
			storage, err := factory.New("wiring")
			if err != nil {
				return fmt.Errorf("failed to create cache: %w", err)
			}

			if err = storage.Set(context.Background(), "key", []byte("v")); err != nil {
				return fmt.Errorf("failed to set key: %w", err)
			}

			return nil
		}),
	)
	if app.Err() != nil {
		t.Fatalf("fx.New failed: %v", app.Err())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if err := app.Stop(ctx); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

// TestModuleSweepIntervalRequired checks that enabling the sweeper without a
// positive interval fails at startup instead of silently never sweeping.
func TestModuleSweepIntervalRequired(t *testing.T) {
	app := fx.New(
		fx.Supply(zap.NewNop()),
		fx.Supply(cachefx.Config{URL: "memory://", SweepInterval: 0}),
		cachefx.Module(cachefx.WithSweeper()),
		appcache.Module(),
	)

	if !errors.Is(app.Err(), cachefx.ErrInvalidFactoryConfig) {
		t.Fatalf("expected ErrInvalidFactoryConfig, got %v", app.Err())
	}
}
