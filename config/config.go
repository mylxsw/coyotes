package config

import "flag"

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

// Channel is the command queue
type Channel struct {
	Name        string      `json:"name"`
	Command     chan string `json:"-"`
	Distinct    bool        `json:"distinct"`
	WorkerCount int         `json:"worker_count"`
}

// Runtime hold global runtime configuration
type Runtime struct {
	Config         Config
	Stoped         chan struct{}
	StopHTTPServer chan struct{}
	StopScheduler  chan struct{}
	Channels       map[string]*Channel
}

var redisAddr = flag.String("redis-host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
var redisPassword = flag.String("redis-password", "", "redis连接密码")
var redisAddrDepressed = flag.String("host", "127.0.0.1:6379", "redis连接地址，必须指定端口(depressed,使用redis-host)")
var redisPasswordDepressed = flag.String("password", "", "redis连接密码(depressed,使用redis-password)")
var httpAddr = flag.String("http-addr", "127.0.0.1:60001", "HTTP监控服务监听地址+端口")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")
var taskMode = flag.Bool("task-mode", true, "是否启用任务模式，默认启用，关闭则不会执行消费")
var colorfulTTY = flag.Bool("colorful-tty", false, "是否启用彩色模式的控制台输出")
var defaultChannel = flag.String("channel-default", "default", "默认channel名称，用于消息队列")

var runtime *Runtime

func init() {
	flag.Parse()

	if *redisAddr == "127.0.0.1:6379" || *redisAddr == "" {
		redisAddr = redisAddrDepressed
	}
	if *redisPassword == "" {
		redisPassword = redisPasswordDepressed
	}

	runtime = &Runtime{
		Config: Config{
			PidFile:    *pidFile,
			Concurrent: *concurrent,
			Redis: RedisConfig{
				Addr:     *redisAddr,
				Password: *redisPassword,
			},
			HTTP: HTTPConfig{
				ListenAddr: *httpAddr,
			},
			TaskMode:         *taskMode,
			ColorfulTTY:      *colorfulTTY,
			DefaultChannel:   *defaultChannel,
			ChannelCacheSize: 20,
		},
		Channels: make(map[string]*Channel),
	}
}

// GetRuntime function return a runtime instance
func GetRuntime() *Runtime {
	return runtime
}
