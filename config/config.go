package config

import (
	"github.com/mylxsw/coyotes/brokers"
	"time"
)

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
	Redis            RedisConfig
	HTTP             HTTPConfig
	Concurrent       int
	PidFile          string
	TaskMode         bool
	ColorfulTTY      bool
	DefaultChannel   string
	ChannelCacheSize int
}

// 进程运行信息
type Info struct {
	StartedAt     time.Time // 开始运行时间
	DealTaskCount int       // 启动以来执行的任务数目
	SuccTaskCount int       // 成功执行的任务数目
	FailTaskCount int       // 执行失败的任务数
}

// Runtime hold global runtime configuration
type Runtime struct {
	Config         Config
	Channels       map[string]*brokers.Channel
	Info           Info
}

var runtime *Runtime

func InitRuntime(
	redisAddr string,
	redisPassword string,
	redisAddrDepressed string,
	redisPasswordDepressed string,
	pidFile string,
	concurrent int,
	redisDB int,
	httpAddr string,
	taskMode bool,
	colorfulTTY bool,
	defaultChannel string,
) *Runtime {

	if redisAddr == "127.0.0.1:6379" || redisAddr == "" {
		redisAddr = redisAddrDepressed
	}
	if redisPassword == "" {
		redisPassword = redisPasswordDepressed
	}

	runtime = &Runtime{
		Config: Config{
			PidFile:    pidFile,
			Concurrent: concurrent,
			Redis: RedisConfig{
				Addr:     redisAddr,
				Password: redisPassword,
				DB:       redisDB,
			},
			HTTP: HTTPConfig{
				ListenAddr: httpAddr,
			},
			TaskMode:         taskMode,
			ColorfulTTY:      colorfulTTY,
			DefaultChannel:   defaultChannel,
			ChannelCacheSize: 20,
		},
		Channels: make(map[string]*brokers.Channel),
		Info:     Info{},
	}

	// 进程启动时间
	runtime.Info.StartedAt = time.Now()

	return runtime
}

// GetRuntime function return a runtime instance
func GetRuntime() *Runtime {
	return runtime
}
