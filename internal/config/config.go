package config

import "time"

type GatewayMode string

const (
	GatewayModePublic  GatewayMode = "public"
	GatewayModePrivate GatewayMode = "private"
)

type Config struct {
	Gateway  Gateway   `yaml:"gateway"`  // gateway config
	HTTP     HTTP      `yaml:"http"`     // http server config
	Database Database  `yaml:"database"` // database config
	FCM      FCMConfig `yaml:"fcm"`      // firebase cloud messaging config
	SSE      SSE       `yaml:"sse"`      // server-sent events config
	Messages Messages  `yaml:"messages"` // messages config
	Cache    Cache     `yaml:"cache"`    // cache (memory or redis) config
	PubSub   PubSub    `yaml:"pubsub"`   // pubsub (memory or redis) config
	JWT      JWT       `yaml:"jwt"`      // jwt config
	OTP      OTP       `yaml:"otp"`      // one-time password config
}

type Gateway struct {
	Mode         GatewayMode `yaml:"mode"          envconfig:"GATEWAY__MODE"`          // gateway mode: public or private
	PrivateToken string      `yaml:"private_token" envconfig:"GATEWAY__PRIVATE_TOKEN"` // device registration token in private mode
}

type HTTP struct {
	Listen  string   `yaml:"listen"  envconfig:"HTTP__LISTEN"`  // listen address
	Proxies []string `yaml:"proxies" envconfig:"HTTP__PROXIES"` // proxies

	API     API     `yaml:"api"`
	OpenAPI OpenAPI `yaml:"openapi"`
}

type API struct {
	Host string `yaml:"host" envconfig:"HTTP__API__HOST"` // public API host
	Path string `yaml:"path" envconfig:"HTTP__API__PATH"` // public API path
}

type OpenAPI struct {
	Enabled bool `yaml:"enabled" envconfig:"HTTP__OPENAPI__ENABLED"` // openapi enabled
}

type Database struct {
	Host     string `yaml:"host"     envconfig:"DATABASE__HOST"`     // database host
	Port     int    `yaml:"port"     envconfig:"DATABASE__PORT"`     // database port
	User     string `yaml:"user"     envconfig:"DATABASE__USER"`     // database user
	Password string `yaml:"password" envconfig:"DATABASE__PASSWORD"` // database password
	Database string `yaml:"database" envconfig:"DATABASE__DATABASE"` // database name
	Timezone string `yaml:"timezone" envconfig:"DATABASE__TIMEZONE"` // database timezone
	Debug    bool   `yaml:"debug"    envconfig:"DATABASE__DEBUG"`    // debug mode

	MaxOpenConns int `yaml:"max_open_conns" envconfig:"DATABASE__MAX_OPEN_CONNS"` // max open connections
	MaxIdleConns int `yaml:"max_idle_conns" envconfig:"DATABASE__MAX_IDLE_CONNS"` // max idle connections
}

type FCMConfig struct {
	CredentialsJSON string `yaml:"credentials_json" envconfig:"FCM__CREDENTIALS_JSON"` // firebase credentials json (public mode only)
	DebounceSeconds uint16 `yaml:"debounce_seconds" envconfig:"FCM__DEBOUNCE_SECONDS"` // push notification debounce (>= 5s)
	TimeoutSeconds  uint16 `yaml:"timeout_seconds"  envconfig:"FCM__TIMEOUT_SECONDS"`  // push notification send timeout
}

type HashingTask struct {
	IntervalSeconds uint16 `yaml:"interval_seconds" envconfig:"TASKS__HASHING__INTERVAL_SECONDS"` // deprecated
}

type SSE struct {
	KeepAlivePeriodSeconds uint16 `yaml:"keep_alive_period_seconds" envconfig:"SSE__KEEP_ALIVE_PERIOD_SECONDS"` // keep alive period in seconds, 0 for no keep alive
}

type Messages struct {
	CacheTTLSeconds        uint16 `yaml:"cache_ttl_seconds"        envconfig:"MESSAGES__CACHE_TTL_SECONDS"`        // cache ttl in seconds
	HashingIntervalSeconds uint16 `yaml:"hashing_interval_seconds" envconfig:"MESSAGES__HASHING_INTERVAL_SECONDS"` // hashing interval in seconds
}

type Cache struct {
	URL string `yaml:"url" envconfig:"CACHE__URL"`
}

type PubSub struct {
	URL        string `yaml:"url"         envconfig:"PUBSUB__URL"`
	BufferSize uint   `yaml:"buffer_size" envconfig:"PUBSUB__BUFFER_SIZE"`
}

type JWT struct {
	Secret string   `yaml:"secret" envconfig:"JWT__SECRET"`
	TTL    Duration `yaml:"ttl"    envconfig:"JWT__TTL"`
	Issuer string   `yaml:"issuer" envconfig:"JWT__ISSUER"`
}

type OTP struct {
	TTL     uint16 `yaml:"ttl"     envconfig:"OTP__TTL"`
	Retries uint8  `yaml:"retries" envconfig:"OTP__RETRIES"`
}

func Default() Config {
	//nolint:exhaustruct,mnd // default values
	return Config{
		Gateway: Gateway{Mode: GatewayModePublic},
		HTTP: HTTP{
			Listen: ":3000",
		},
		Database: Database{
			Host:     "localhost",
			Port:     3306,
			User:     "sms",
			Password: "sms",
			Database: "sms",
			Timezone: "UTC",
		},
		FCM: FCMConfig{
			CredentialsJSON: "",
		},
		SSE: SSE{
			KeepAlivePeriodSeconds: 15,
		},
		Messages: Messages{
			CacheTTLSeconds:        300, // 5 minutes
			HashingIntervalSeconds: 60,
		},
		Cache: Cache{
			URL: "memory://",
		},
		PubSub: PubSub{
			URL:        "memory://",
			BufferSize: 128,
		},
		JWT: JWT{
			TTL:    Duration(time.Hour * 24),
			Issuer: "sms-gate.app",
		},
		OTP: OTP{
			TTL:     300,
			Retries: 3,
		},
	}
}
