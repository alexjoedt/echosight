package app

import "fmt"

type Config struct {
	// Loglevel is the loglevel of the app.
	//
	// Vaild values: `debug`, `info`, `warn`, `error`
	//
	// default: `info`
	LogLevel string `json:"logLevel" toml:"logLevel" yaml:"logLevel" env:"LOG_LEVEL"`

	// Environment where the server runs, production or development.
	//
	// Valid values: `development`, `dev`, `production`, `prod`
	//
	// default: `prod`
	Environment string `json:"environment" toml:"environment" yaml:"environment" env:"ENVIRONMENT"`

	// Secret is used to encrypt and decrypt passwords and other sensitive data.
	// It must be 32 bytes long
	//
	// `required`
	Secret string `json:"secret" toml:"secret" yaml:"secret" env:"BLAZE_SECRET"`

	HTTP struct {
		// The port specifies the HTTP location where the server's API is accessible.
		//
		// default: `8081`
		Port string `json:"port" toml:"port" yaml:"port" env:"HTTP_PORT"`

		// TrustedOrigins, an array of trusted origins, because of CORS.
		// The pulic address of the client should be entered.
		//
		// For the environment variable mulitple values can be comma seperated
		TrustedOrigins []string `json:"trustedOrigins" toml:"trustedOrigins" yaml:"trustedOrigins" env:"HTTP_TRUSTED_ORIGINS"`

		// Limiter is a HTTP RateLimiter to prevent attacks
		Limiter struct {
			// This indicates the maximum number of requests a client is allowed to send within a specific time period.
			RateLimit float64 `json:"rateLimit" toml:"rateLimit" yaml:"rateLimit" env:"LIMITER_RATE_LIMIT"`

			// Burst refers to the maximum number of requests that are temporarily allowed above the normal RateLimit.
			// This allows a system to handle occasional spikes in traffic.
			Burst int `json:"burst" toml:"burst" yaml:"burst" env:"LIMITER_BURST"`

			// Enabled enables the Limiter.
			//
			// default: `true`
			Enabled bool `json:"enabled" toml:"enabled" yaml:"enabled" env:"LIMITER_ENABLED"`
		} `json:"limiter" toml:"limiter" yaml:"limiter"`
	} `json:"http" toml:"http" yaml:"http"`

	// PostgreSQL Database connection information
	Postgres struct {
		// required
		Host string `json:"host" toml:"host" yaml:"host" env:"PSQL_HOST"`
		// required
		Port string `json:"port" toml:"port" yaml:"port" env:"PSQL_PORT"`
		SSL  bool   `json:"ssl" toml:"ssl" yaml:"ssl" env:"PSQL_SSL"`
		// required
		User string `json:"user" toml:"user" yaml:"user" env:"PSQL_USER"`
		// required
		Password string `json:"password" toml:"password" yaml:"password" env:"PSQL_PASSWORD"`
		Database string `json:"-" toml:"-" yaml:"-"`
	} `json:"postgres" toml:"postgres" yaml:"postgres"`

	// InfluxDB connection information
	//
	// Influx is used for collecting, metrics from the host and detectors.
	InfluxDB struct {
		// required
		Host string `json:"host" toml:"host" yaml:"host" env:"INFLUXDB_HOST"`
		// required
		Port string `json:"port" toml:"port" yaml:"port" env:"INFLUXDB_PORT"`
		SSL  bool   `json:"ssl" toml:"ssl" yaml:"ssl" env:"INFLUXDB_SSL"`
		// required
		Token string `json:"token" toml:"token" yaml:"token" env:"INFLUXDB_TOKEN"`
	} `json:"influxdb" toml:"influxdb" yaml:"influxdb"`

	// Redis connection information
	//
	// Redis is used for messaging, cache and queue.
	Redis struct {
		// required
		Host string `json:"host" toml:"host" yaml:"host" env:"REDIS_HOST"`
		// required
		Port string `json:"port" toml:"port" yaml:"port" env:"REDIS_PORT"`
		// required
		Password string `json:"password" toml:"password" yaml:"password" env:"REDIS_PASSWORD"`
	} `json:"redis" toml:"redis" yaml:"redis"`

	Cache struct {
		// valid values: `local`, `redis`
		//
		// default `local`
		CacheType CacheType `json:"cacheType" toml:"cacheType" yaml:"cacheType" env:"CACHE_TYPE"`

		// Time To Live
		//
		// currently this has no effect
		TTL Duration `json:"ttl" toml:"ttl" yaml:"ttl" env:"CACHE_TTL"`
	} `json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty"`

	// Mailserver connection information
	SMTP struct {
		Host     string `json:"host" toml:"host" yaml:"host" env:"SMTP_HOST"`
		Port     string `json:"port" toml:"port" yaml:"port" env:"SMTP_PORT"`
		Username string `json:"username" toml:"username" yaml:"username" env:"SMTP_USER"`
		Password string `json:"password" toml:"password" yaml:"password" env:"SMTP_PASSWORD"`

		// Sender email address
		Sender string `json:"sender" toml:"sender" yaml:"sender" env:"SMTP_SENDER"`
	} `json:"smtp" toml:"smtp" yaml:"smtp"`
}

func (c *Config) postgresDSN() string {
	database := "echosight"
	if c.isDev() {
		database = fmt.Sprintf("%s_dev", database)
	}

	sslmode := "disable"
	if c.Postgres.SSL {
		sslmode = "enable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		database,
		sslmode,
	)
}

func (c *Config) InfluxURL() string {
	scheme := "http"
	if c.Postgres.SSL {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s:%s",
		scheme,
		c.InfluxDB.Host,
		c.InfluxDB.Port,
	)
}
