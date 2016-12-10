package config

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type HttpConfig struct {
	ListenAddr string
}

type Config struct {
	Redis          RedisConfig
	Http           HttpConfig
	Concurrent     int
	PidFile        string
	TaskMode       bool
	ColorfulTTY    bool
	DefaultChannel string
}

type Channel struct {
	Name    string
	Command chan string
}

type Runtime struct {
	Config   Config
	Stoped   chan struct{}
	Channels map[string]*Channel
}
