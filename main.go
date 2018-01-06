package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jpillora/overseer/fetcher"

	"github.com/jpillora/overseer"

	"github.com/mylxsw/coyotes/backend"
	"github.com/mylxsw/coyotes/backend/mysql"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/pidfile"
	"github.com/mylxsw/coyotes/scheduler"

	sysRuntime "runtime"

	_ "github.com/go-sql-driver/mysql"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	server "github.com/mylxsw/coyotes/http"
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
	logFilename            string
	debugMode              bool
	daemonize              bool
	backendStorage         string
	backendKeepDays        int
	fetchUpdateURL         string
)

var BuildID = "0"

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
	flag.StringVar(&logFilename, "log-file", "", "日志文件存储路径，默认为空，直接输出到标准输出")
	flag.BoolVar(&debugMode, "debug", false, "日志输出级别，默认为false，如果为true，则输出debug日志")
	flag.BoolVar(&daemonize, "daemonize", false, "守护进程模式，模式为false")
	flag.StringVar(&backendStorage, "backend-storage", "", "后端存储方式，用于存储任务执行结果，默认不存储")
	flag.IntVar(&backendKeepDays, "backend-keep-days", 0, "后端存储历史保留天数，0为永久保留")
	flag.StringVar(&fetchUpdateURL, "update-check-url", "https://aicode.cc/open-api/coyotes/update/coyotes-%s-%s", "自动更新检查地址")

	flag.Parse()

	// 如果是守护进程模式，则创建子进程，退出父进程
	if daemonize && os.Getppid() != 1 {
		binary, err := exec.LookPath(os.Args[0])
		if err != nil {
			fmt.Println("failed to lookup binary:", err)
			os.Exit(2)
		}
		_, err = os.StartProcess(binary, os.Args, &os.ProcAttr{Dir: "", Env: nil, Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: nil})
		if err != nil {
			fmt.Println("failed to start process:", err)
			os.Exit(2)
		}

		os.Exit(0)
	}

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
		logFilename,
		debugMode,
		backendStorage,
		backendKeepDays,
	)

	if os.Getuid() == 0 {
		fmt.Println(console.ColorfulText(
			console.TextYellow,
			"\n当前以root用户执行，使用root权限执行可能会造成严重的安全问题，建议使用非root用户执行\n",
		))
	}

	overseer.Run(overseer.Config{
		Program: mainProcess,
		Address: runtime.Config.HTTP.ListenAddr,
		Debug:   runtime.Config.DebugMode,
		Fetcher: &fetcher.HTTP{
			URL:      fmt.Sprintf(fetchUpdateURL, sysRuntime.GOOS, sysRuntime.GOARCH),
			Interval: 5 * time.Second,
		},
	})
}

func mainProcess(state overseer.State) {

	runtime := config.GetRuntime()
	runtime.BuildID = BuildID

	// 初始化日志输出
	// 指定日志文件时，使用日志文件输出，否则输出到标准输出
	if runtime.Config.LogFilename != "" {
		logFile, err := os.OpenFile(runtime.Config.LogFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			fmt.Printf("open log file %s failed: %v\n", runtime.Config.LogFilename, err)
			os.Exit(2)
		}
		defer logFile.Close()

		log.InitLogger(logFile, debugMode, "coyotes#"+BuildID)
	} else {
		log.InitLogger(os.Stdout, debugMode, "coyotes#"+BuildID)
	}

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
		fmt.Println(console.ColorfulText(console.TextGreen, config.WelcomeMessage()))
	}

	log.Debug("redis addr: %s/%d", runtime.Config.Redis.Addr, runtime.Config.Redis.DB)
	log.Debug("process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	ctx, cancel := context.WithCancel(context.Background())
	// signal.InitSignalReceiver(ctx, cancel)

	go func() {
		<-state.GracefulShutdown
		cancel()
	}()

	// 初始化所有channel
	scheduler.InitChannels()

	// 初始化后端存储
	backendStorage := runtime.Config.BackendStorage
	if backendStorage != "" && strings.HasPrefix(backendStorage, "mysql:") {
		dataSource := backendStorage[6:]

		mysql.Register("mysql", dataSource)
		mysql.InitTableForMySQL(dataSource)

		// 自动清理过期的后端存储日志
		if runtime.Config.BackendKeepDays > 0 {
			go func() {
				for _ = range time.Tick(5 * time.Minute) {
					beforeTime := time.Now().AddDate(0, 0, -runtime.Config.BackendKeepDays)
					if driver := backend.Default(); driver != nil {
						affectRows, err := driver.ClearExpired(beforeTime)
						if err != nil {
							log.Error("backend clear hisories failed: %v", err)
							return
						}

						log.Debug("backend clear histories, affected %d rows", affectRows)
					}
				}
			}()
		}
	}

	var wg sync.WaitGroup

	// 启动http server
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.StartHTTPServer(state.Listener)
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
