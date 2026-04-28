package smpp

import (
	"context"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"smpp",
		fx.Provide(
			NewServer,
			NewHandler,
			NewWebhookHandler,
		),
		fx.Invoke(Register),
	)
}

type StartParams struct {
	fx.In

	LC     fx.Lifecycle
	Logger *zap.Logger

	Server  *Server
	Handler *Handler
	Webhook *WebhookHandler
}

func Register(p StartParams) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Wire up Handler with Server and WebhookHandler for metrics aggregation
	p.Handler.SetServer(p.Server, p.Webhook)

	p.LC.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := p.Server.Start(ctx); err != nil {
					p.Logger.Error("SMPP server error", zap.Error(err))
				}
			}()

			p.Logger.Info("SMPP server started")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			wg.Wait()
			p.Logger.Info("SMPP server stopped")

			return nil
		},
	})

	return nil
}
