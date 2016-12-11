package main

import (
	"flag"
	"fmt"
	"os"

	"sync"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
	"github.com/mylxsw/task-runner/pidfile"
	"github.com/mylxsw/task-runner/signal"
)

var redisAddr = flag.String("host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
var redisPassword = flag.String("password", "", "redis连接密码")
var httpAddr = flag.String("http-addr", "127.0.0.1:60001", "HTTP监控服务监听地址+端口")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")
var taskMode = flag.Bool("task-mode", true, "是否启用任务模式，默认启用，关闭则不会执行消费")
var colorfulTTY = flag.Bool("colorful-tty", false, "是否启用彩色模式的控制台输出")
var defaultChannel = flag.String("channel-default", "default", "默认channel名称，用于消息队列")

func main() {

	flag.Parse()

	runtime := &config.Runtime{
		Config: config.Config{
			PidFile:    *pidFile,
			Concurrent: *concurrent,
			Redis: config.RedisConfig{
				Addr:     *redisAddr,
				Password: *redisPassword,
			},
			HTTP: config.HTTPConfig{
				ListenAddr: *httpAddr,
			},
			TaskMode:       *taskMode,
			ColorfulTTY:    *colorfulTTY,
			DefaultChannel: *defaultChannel,
		},
		Channels: map[string]*config.Channel{
			*defaultChannel: &config.Channel{
				Name:     *defaultChannel,
				Command:  make(chan string, 20),
				Distinct: true,
			},
			"biz": &config.Channel{
				Name:     "biz",
				Command:  make(chan string, 20),
				Distinct: true,
			},
			"normal": &config.Channel{
				Name:     "normal",
				Command:  make(chan string, 20),
				Distinct: false,
			},
		},
	}

	// 用于向所有channel发送程序退出信号
	runtime.Stoped = make(chan struct{}, len(runtime.Channels))

	// 创建进程pid文件
	pid, err := pidfile.New(*pidFile)
	if err != nil {
		log.Error("Failed to create pidfile: %v", err)
		os.Exit(2)
	}
	defer pid.Remove()

	if *colorfulTTY {
		fmt.Println(console.ColorfulText(runtime, console.TextCyan, welcomeMessage(runtime)))
	}

	log.Debug("The redis addr: %s", runtime.Config.Redis.Addr)
	log.Debug("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver(runtime)

	go startHTTPServer(runtime)

	var wg sync.WaitGroup

	for index := range runtime.Channels {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()
			startTaskRunner(runtime, runtime.Channels[i])
		}(index)
	}

	wg.Wait()
}
