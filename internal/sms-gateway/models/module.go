package models

import (
	"github.com/capcom6/go-infra-fx/db"
)

//nolint:gochecknoinits // framework-specific
func init() {
	db.RegisterMigration(Migrate)
	db.RegisterGoose(migrations)
}
