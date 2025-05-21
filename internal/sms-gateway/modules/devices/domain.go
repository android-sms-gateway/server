package devices

import "github.com/capcom6/go-helpers/anys"

type Settings struct {
	Encryption  *SettingsEncryption  `json:"encryption,omitempty"`
	Gateway     *SettingsGateway     `json:"gateway,omitempty"`
	Messages    *SettingsMessages    `json:"messages,omitempty"`
	Localserver *SettingsLocalserver `json:"localserver,omitempty"`
	Ping        *SettingsPing        `json:"ping,omitempty"`
	Logs        *SettingsLogs        `json:"logs,omitempty"`
	Webhooks    *SettingsWebhooks    `json:"webhooks,omitempty"`
}

type SettingsEncryption struct {
	Passphrase *anys.Optional[string] `json:"passphrase,omitempty"`
}

type SettingsGateway struct {
	CloudURL     *anys.Optional[string] `json:"cloud_url,omitempty"`
	PrivateToken *anys.Optional[string] `json:"private_token,omitempty"`
}

type SettingsMessages struct {
	SendIntervalMin  *anys.Optional[int]    `json:"send_interval_min,omitempty"`
	SendIntervalMax  *anys.Optional[int]    `json:"send_interval_max,omitempty"`
	LimitPeriod      *anys.Optional[string] `json:"limit_period,omitempty"`
	LimitValue       *anys.Optional[int]    `json:"limit_limit_value,omitempty"`
	SimSelectionMode *anys.Optional[string] `json:"sim_selection_mode,omitempty"`
	LogLifetimeDays  *anys.Optional[int]    `json:"log_lifetime_days,omitempty"`
}

type SettingsLocalserver struct {
	PORT *anys.Optional[int] `json:"PORT,omitempty"`
}

type SettingsPing struct {
	IntervalSeconds *anys.Optional[int] `json:"interval_seconds,omitempty"`
}

type SettingsLogs struct {
	LifetimeDays *anys.Optional[int] `json:"lifetime_days,omitempty"`
}

type SettingsWebhooks struct {
	InternetRequired *anys.Optional[bool]   `json:"internet_required,omitempty"`
	RetryCount       *anys.Optional[int]    `json:"retry_count,omitempty"`
	SigningKey       *anys.Optional[string] `json:"signing_key,omitempty"`
}
