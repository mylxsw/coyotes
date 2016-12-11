package config

// RedisConfig hold redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// HTTPConfig hold http configuration
type HTTPConfig struct {
	ListenAddr string
}

// Config hold all the configuration
type Config struct {
	Redis          RedisConfig
	HTTP           HTTPConfig
	Concurrent     int
	PidFile        string
	TaskMode       bool
	ColorfulTTY    bool
	DefaultChannel string
}

// Channel is the command queue
type Channel struct {
	Name     string
	Command  chan string
	Distinct bool
}

// Runtime hold global runtime configuration
type Runtime struct {
	Config   Config
	Stoped   chan struct{}
	Channels map[string]*Channel
}
