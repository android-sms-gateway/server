package config

type GatewayMode string

const (
	GatewayModePublic  GatewayMode = "public"
	GatewayModePrivate GatewayMode = "private"
)

type Config struct {
	Gateway  Gateway   `koanf:"gateway"`  // gateway config
	HTTP     HTTP      `koanf:"http"`     // http server config
	Database Database  `koanf:"database"` // database config
	FCM      FCMConfig `koanf:"fcm"`      // firebase cloud messaging config
	Tasks    Tasks     `koanf:"tasks"`    // tasks config
	SSE      SSE       `koanf:"sse"`      // server-sent events config
}

type Gateway struct {
	Mode         GatewayMode `koanf:"mode"`          // gateway mode: public or private
	PrivateToken string      `koanf:"private_token"` // device registration token in private mode
}

type HTTP struct {
	Listen  string   `koanf:"listen"`  // listen address
	Proxies []string `koanf:"proxies"` // proxies

	API     API     `koanf:"api"`
	OpenAPI OpenAPI `koanf:"openapi"`
}

type API struct {
	Host string `koanf:"host"` // public API host
	Path string `koanf:"path"` // public API path
}

type OpenAPI struct {
	Enabled bool `koanf:"enabled"` // openapi enabled
}

type Database struct {
	Dialect  string `koanf:"dialect"`  // database dialect
	Host     string `koanf:"host"`     // database host
	Port     int    `koanf:"port"`     // database port
	User     string `koanf:"user"`     // database user
	Password string `koanf:"password"` // database password
	Database string `koanf:"database"` // database name
	Timezone string `koanf:"timezone"` // database timezone
	Debug    bool   `koanf:"debug"`    // debug mode

	MaxOpenConns int `koanf:"max_open_conns"` // max open connections
	MaxIdleConns int `koanf:"max_idle_conns"` // max idle connections
}

type FCMConfig struct {
	CredentialsJSON string `koanf:"credentials_json"` // firebase credentials json (public mode only)
	DebounceSeconds uint16 `koanf:"debounce_seconds"` // push notification debounce (>= 5s)
	TimeoutSeconds  uint16 `koanf:"timeout_seconds"`  // push notification send timeout
}

type Tasks struct {
	Hashing HashingTask `koanf:"hashing"`
}

type HashingTask struct {
	IntervalSeconds uint16 `koanf:"interval_seconds"` // hashing interval in seconds
}

type SSE struct {
	KeepAlivePeriodSeconds uint16 `koanf:"keep_alive_period_seconds"` // keep alive period in seconds, 0 for no keep alive
}

var defaultConfig = Config{
	Gateway: Gateway{Mode: GatewayModePublic},
	HTTP: HTTP{
		Listen: ":3000",
	},
	Database: Database{
		Dialect:  "mysql",
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
	Tasks: Tasks{
		Hashing: HashingTask{
			IntervalSeconds: uint16(15 * 60),
		},
	},
	SSE: SSE{
		KeepAlivePeriodSeconds: 15,
	},
}
