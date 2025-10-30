package tasks

import (
	"github.com/android-sms-gateway/server/internal/worker/tasks/devices"
	"github.com/android-sms-gateway/server/internal/worker/tasks/messages"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"tasks",
		logger.WithNamedLogger("tasks"),
		messages.Module(),
		devices.Module(),
	)
}
