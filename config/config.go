package config

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type HttpConfig struct {
	ListenAddr string
}

type Runtime struct {
	Redis           RedisConfig
	Http            HttpConfig
	Concurrent      int
	PidFile         string
	StopRunning     bool
	StopRunningChan chan struct{}
	Command         chan string
	TaskMode        bool
	ColorfulTTY     bool
}
