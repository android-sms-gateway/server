package smsgateway

import (
	"context"
	"sync"

	appconfig "github.com/android-sms-gateway/server/internal/config"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/cleaner"
	appdb "github.com/android-sms-gateway/server/internal/sms-gateway/modules/db"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/health"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/metrics"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/settings"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/webhooks"
	"github.com/capcom6/go-infra-fx/cli"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/capcom6/go-infra-fx/http"
	"github.com/capcom6/go-infra-fx/logger"
	"github.com/capcom6/go-infra-fx/validator"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Module = fx.Module(
	"server",
	logger.Module,
	appconfig.Module,
	appdb.Module,
	http.Module,
	validator.Module,
	handlers.Module,
	auth.Module,
	push.Module,
	db.Module,
	messages.Module,
	health.Module,
	webhooks.Module,
	settings.Module,
	devices.Module,
	metrics.Module,
	cleaner.Module,
)

func Run() {
	cli.DefaultCommand = "start"
	fx.New(
		cli.GetModule(),
		Module,
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			logOption := fxevent.ZapLogger{Logger: logger}
			logOption.UseLogLevel(zapcore.DebugLevel)
			return &logOption
		}),
	).Run()
}

type StartParams struct {
	fx.In

	LC     fx.Lifecycle
	Logger *zap.Logger
	Shut   fx.Shutdowner

	Server          *http.Server
	MessagesService *messages.Service
	PushService     *push.Service
	CleanerService  *cleaner.Service
}

func Start(p StartParams) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	p.LC.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			p.MessagesService.RunBackgroundTasks(ctx, wg)

			wg.Add(1)
			go func() {
				defer wg.Done()
				p.PushService.Run(ctx)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := p.Server.Start(); err != nil {
					p.Logger.Error("Error starting server", zap.Error(err))
					_ = p.Shut.Shutdown()
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				p.CleanerService.Run(ctx)
			}()

			p.Logger.Info("Service started")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			_ = p.Server.Stop(ctx)
			wg.Wait()

			p.Logger.Info("Service stopped")

			return nil
		},
	})

	return nil
}

func init() {
	cli.Register("start", Start)
}
