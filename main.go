package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/pidfile"
	"github.com/mylxsw/coyotes/scheduler"
	"github.com/mylxsw/coyotes/signal"

	broker "github.com/mylxsw/coyotes/brokers/redis"
	server "github.com/mylxsw/coyotes/http"
	"context"
	"sync"
)

var (
	redisAddr              string
	redisPassword          string
	redisDB                int
	redisAddrDepressed     string
	redisPasswordDepressed string
	httpAddr               string
	pidFile                string
	concurrent             int
	taskMode               bool
	colorfulTTY            bool
	defaultChannel         string
)

func main() {
	flag.Usage = func() {
		fmt.Println(config.WelcomeMessageStr)
		fmt.Print("Options:\n\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&redisAddr, "redis-host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
	flag.StringVar(&redisPassword, "redis-password", "", "redis连接密码")
	flag.IntVar(&redisDB, "redis-db", 0, "redis默认数据库0-15")
	flag.StringVar(&redisAddrDepressed, "host", "127.0.0.1:6379", "redis连接地址，必须指定端口(depressed,使用redis-host)")
	flag.StringVar(&redisPasswordDepressed, "password", "", "redis连接密码(depressed,使用redis-password)")
	flag.StringVar(&httpAddr, "http-addr", "127.0.0.1:60001", "HTTP监控服务监听地址+端口")
	flag.StringVar(&pidFile, "pidfile", "", "pid文件路径，默认为空，不使用")
	flag.IntVar(&concurrent, "concurrent", 5, "并发执行线程数")
	flag.BoolVar(&taskMode, "task-mode", true, "是否启用任务模式，默认启用，关闭则不会执行消费")
	flag.BoolVar(&colorfulTTY, "colorful-tty", false, "是否启用彩色模式的控制台输出")
	flag.StringVar(&defaultChannel, "channel-default", "default", "默认channel名称，用于消息队列")

	flag.Parse()

	runtime := config.InitRuntime(
		redisAddr,
		redisPassword,
		redisAddrDepressed,
		redisPasswordDepressed,
		pidFile,
		concurrent,
		redisDB,
		httpAddr,
		taskMode,
		colorfulTTY,
		defaultChannel,
	)

	// 创建进程pid文件
	if runtime.Config.PidFile != "" {
		pid, err := pidfile.New(runtime.Config.PidFile)
		if err != nil {
			log.Error("failed to create pidfile: %v", err)
			os.Exit(2)
		}
		defer pid.Remove()
	}

	if runtime.Config.ColorfulTTY {
		fmt.Println(console.ColorfulText(console.TextCyan, config.WelcomeMessage()))
	}

	log.Debug("redis addr: %s/%d", runtime.Config.Redis.Addr, runtime.Config.Redis.DB)
	log.Debug("process ID: %d", os.Getpid())


	// 信号处理程序，接收退出信号，平滑退出进程
	ctx, cancel := context.WithCancel(context.Background())
	signal.InitSignalReceiver(ctx, cancel)

	// 初始化所有channel
	scheduler.InitChannels()

	var wg sync.WaitGroup

	// 启动http server
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.StartHTTPServer(ctx)
	}()

	// 启动待执行任务转移任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		broker.TransferPrepareTask(ctx)
	}()

	// 启动延迟任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		broker.StartDelayTaskLifeCycle(ctx)
	}()

	// 启动任务调度器
	scheduler.Schedule(ctx)

	wg.Wait()
	log.Debug("all stoped.")
}
