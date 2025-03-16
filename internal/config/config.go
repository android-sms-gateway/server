package config

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
	Tasks    Tasks     `yaml:"tasks"`    // tasks config
}

type Gateway struct {
	Mode         GatewayMode `yaml:"mode"          envconfig:"GATEWAY__MODE"`          // gateway mode: public or private
	PrivateToken string      `yaml:"private_token" envconfig:"GATEWAY__PRIVATE_TOKEN"` // device registration token in private mode
}

type HTTP struct {
	Listen  string   `yaml:"listen" envconfig:"HTTP__LISTEN"`   // listen address
	Proxies []string `yaml:"proxies" envconfig:"HTTP__PROXIES"` // proxies
}

type Database struct {
	Dialect  string `yaml:"dialect"  envconfig:"DATABASE__DIALECT"`  // database dialect
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

type Tasks struct {
	Hashing HashingTask `yaml:"hashing"`
}

type HashingTask struct {
	IntervalSeconds uint16 `yaml:"interval_seconds" envconfig:"TASKS__HASHING__INTERVAL_SECONDS"` // hashing interval in seconds
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
}
