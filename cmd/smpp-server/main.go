package main

import (
	"os"

	"github.com/android-sms-gateway/server/internal/config"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/smpp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	cmdStart = "start"
)

func main() {
	args := os.Args[1:]
	cmd := cmdStart
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case cmdStart:
		run()
	}
}

func run() {
	cfg := config.Default()

	fx.New(
		fx.NopLogger,
		fx.Provide(
			func() smpp.Config {
				return smpp.Config{
					BindAddress:    cfg.SMPP.BindAddress,
					TLSBindAddress: cfg.SMPP.TLSBindAddress,
					TLSCert:        cfg.SMPP.TLSCert,
					TLSKey:         cfg.SMPP.TLSKey,
					APIBaseURL:     cfg.SMPP.APIBaseURL,
					WebhookBaseURL: cfg.SMPP.WebhookBaseURL,
				}
			},
		),
		smpp.Module(),
		fx.Invoke(func(p smpp.StartParams) error {
			return nil
		}),
	).Run()

	if zap.L() != nil {
		zap.L().Info("SMPP server shutdown complete")
	}
}
