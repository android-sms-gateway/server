package smsgateway

import (
	"bitbucket.org/capcom6/smsgatewaybackend/internal/smsgateway/handlers"
	"bitbucket.org/capcom6/smsgatewaybackend/internal/smsgateway/models"
	_ "bitbucket.org/capcom6/smsgatewaybackend/internal/smsgateway/tasks"
	microbase "bitbucket.org/soft-c/gomicrobase"
)

func init() {
	microbase.RegisterMigration(models.Migrate)
	microbase.RegisterHandlers(handlers.Register)
}
