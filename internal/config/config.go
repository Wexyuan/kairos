package config

// Config holds the application configuration.
type Config struct {
	App      App      `mapstructure:"app"`      // application settings
	Logger   Logger   `mapstructure:"logger"`   // logger settings
	Database Database `mapstructure:"database"` // database settings
	Cache    Cache    `mapstructure:"cache"`    // cache settings
}

// App holds the application settings.
type App struct {
	Name          string `mapstructure:"name"`           // application name
	SnowflakeNode int64  `mapstructure:"snowflake_node"` // snowflake worker node (0-1023)
}

// Logger holds the logger settings.
type Logger struct {
	Level string `mapstructure:"level"` // minimum log level (debug/info/warn/error)
	File  string `mapstructure:"file"`  // log file path, empty disables file output
}

// Database holds the database settings.
type Database struct {
	DSN             string `mapstructure:"dsn"`               // data source name, a sqlite file path
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    // max open connections (default to 10)
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    // max idle connections (default to 2)
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // max connection lifetime in minutes (default to 30)
}

// Cache holds the cache settings.
type Cache struct {
	TTL int `mapstructure:"ttl"` // expiration in seconds (default to 300)
}
