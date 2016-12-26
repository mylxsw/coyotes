package config

import (
	"flag"
	"fmt"
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
	Name        string
	Command     chan string
	Distinct    bool
	WorkerCount int
}

// Runtime hold global runtime configuration
type Runtime struct {
	Config         Config
	Stoped         chan struct{}
	StopHTTPServer chan struct{}
	StopScheduler  chan struct{}
	Channels       map[string]*Channel
}

var redisAddr = flag.String("host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
var redisPassword = flag.String("password", "", "redis连接密码")
var httpAddr = flag.String("http-addr", "127.0.0.1:60001", "HTTP监控服务监听地址+端口")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")
var taskMode = flag.Bool("task-mode", true, "是否启用任务模式，默认启用，关闭则不会执行消费")
var colorfulTTY = flag.Bool("colorful-tty", false, "是否启用彩色模式的控制台输出")
var defaultChannel = flag.String("channel-default", "default", "默认channel名称，用于消息队列")

const channelCacheSize = 20

var runtime *Runtime

func init() {
	flag.Parse()

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
			TaskMode:       *taskMode,
			ColorfulTTY:    *colorfulTTY,
			DefaultChannel: *defaultChannel,
		},
		Channels: map[string]*Channel{
			*defaultChannel: &Channel{
				Name:        *defaultChannel,
				Command:     make(chan string, channelCacheSize),
				Distinct:    true,
				WorkerCount: *concurrent,
			},
			"biz": &Channel{
				Name:        "biz",
				Command:     make(chan string, channelCacheSize),
				Distinct:    true,
				WorkerCount: *concurrent,
			},
			"cron": &Channel{
				Name:        "cron",
				Command:     make(chan string, channelCacheSize),
				Distinct:    true,
				WorkerCount: *concurrent,
			},
		},
	}

	// 用于向所有channel发送程序退出信号
	runtime.Stoped = make(chan struct{}, len(runtime.Channels))
	runtime.StopHTTPServer = make(chan struct{})
	runtime.StopScheduler = make(chan struct{})
}

// GetRuntime function return a runtime instance
func GetRuntime() *Runtime {
	return runtime
}

// NewChannel function create a new channel for task queue
func NewChannel(name string, distinct bool, workerCount int) (*Channel, error) {
	if name == "" {
		return nil, fmt.Errorf("队列名称不能为空")
	}
	if _, ok := runtime.Channels[name]; ok {
		return nil, fmt.Errorf("任务队列 %s 已经存在", name)
	}

	channel := &Channel{
		Name:        name,
		Command:     make(chan string, channelCacheSize),
		Distinct:    distinct,
		WorkerCount: workerCount,
	}

	if workerCount == 0 {
		channel.WorkerCount = *concurrent
	}

	runtime.Channels[name] = channel
	return channel, nil
}
