package config

import (
	"time"

	"github.com/android-sms-gateway/server/internal/config"
)

type Config struct {
	Tasks    Tasks           `yaml:"tasks"`
	Database config.Database `yaml:"database"`
	HTTP     config.HTTP     `yaml:"http"`
}

type Tasks struct {
	MessagesHashing MessagesHashing `yaml:"messages_hashing"`
	MessagesCleanup MessagesCleanup `yaml:"messages_cleanup"`
	DevicesCleanup  DevicesCleanup  `yaml:"devices_cleanup"`
}
type MessagesHashing struct {
	Interval Duration `yaml:"interval" envconfig:"TASKS__MESSAGES_HASHING__INTERVAL"`
}

type MessagesCleanup struct {
	Interval Duration `yaml:"interval" envconfig:"TASKS__MESSAGES_CLEANUP__INTERVAL"`
	MaxAge   Duration `yaml:"max_age"  envconfig:"TASKS__MESSAGES_CLEANUP__MAX_AGE"`
}

type DevicesCleanup struct {
	Interval Duration `yaml:"interval" envconfig:"TASKS__DEVICES_CLEANUP__INTERVAL"`
	MaxAge   Duration `yaml:"max_age"  envconfig:"TASKS__DEVICES_CLEANUP__MAX_AGE"`
}

func Default() Config {
	return Config{
		Tasks: Tasks{
			MessagesHashing: MessagesHashing{
				Interval: Duration(7 * 24 * time.Hour),
			},
			MessagesCleanup: MessagesCleanup{
				Interval: Duration(24 * time.Hour),
				MaxAge:   Duration(30 * 24 * time.Hour),
			},
			DevicesCleanup: DevicesCleanup{
				Interval: Duration(24 * time.Hour),
				MaxAge:   Duration(365 * 24 * time.Hour),
			},
		},
		Database: config.Database{
			Host:         "localhost",
			Port:         3306,
			User:         "sms",
			Password:     "sms",
			Database:     "sms",
			Timezone:     "UTC",
			Debug:        false,
			MaxOpenConns: 0,
			MaxIdleConns: 0,
		},
		HTTP: config.HTTP{
			Listen:  "127.0.0.1:3000",
			Proxies: []string{},
		},
	}
}
