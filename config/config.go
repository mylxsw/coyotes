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
	Redis       RedisConfig
	Http        HttpConfig
	Concurrent  int
	PidFile     string
	TaskMode    bool
	ColorfulTTY bool
}

type Runtime struct {
	Config          Config
	StopRunning     bool
	StopRunningChan chan struct{}
	Command         chan string
}
