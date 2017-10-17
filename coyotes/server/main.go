package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/pidfile"
	"github.com/mylxsw/coyotes/scheduler"
	"github.com/mylxsw/coyotes/signal"
	"github.com/urfave/cli"

	"os/exec"

	broker "github.com/mylxsw/coyotes/brokers/redis"
	server "github.com/mylxsw/coyotes/http"
)

func main() {

	app := cli.NewApp()
	app.Name = "coyotes"
	app.Version = "2.0"
	app.Usage = "一款轻量级的分布式任务执行队列"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "mylxsw",
			Email: "mylxsw@aicode.cc",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "pidfile",
			Value: "",
			Usage: "pid文件路径，默认为空，不使用",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "日志输出级别，默认为false，如果为true，则输出debug日志",
		},
		cli.BoolFlag{
			Name:  "daemonize, d",
			Usage: "守护进程模式，模式为false",
		},
		cli.BoolFlag{
			Name:  "colorful-tty",
			Usage: "是否启用彩色模式的控制台输出",
		},
		cli.StringFlag{
			Name:  "redis-host",
			Value: "127.0.0.1:6379",
			Usage: "redis连接地址，必须指定端口",
		},
		cli.StringFlag{
			Name:  "redis-password",
			Value: "",
			Usage: "redis连接密码",
		},
		cli.IntFlag{
			Name:  "redis-db",
			Value: 0,
			Usage: "redis默认数据库0-15",
		},
		cli.IntFlag{
			Name:  "concurrent, c",
			Value: 5,
			Usage: "并发执行线程数",
		},
		cli.StringFlag{
			Name:  "channel-default",
			Value: "default",
			Usage: "默认channel名称，用于消息队列",
		},
		cli.StringFlag{
			Name:  "log-file",
			Value: "",
			Usage: "日志文件存储路径，默认为空，直接输出到标准输出",
		},
		cli.StringFlag{
			Name:  "http-addr",
			Value: "127.0.0.1:60001",
			Usage: "HTTP监控服务监听地址+端口",
		},
	}

	app.Action = func(c *cli.Context) error {
		daemonize := c.Bool("daemonize")
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
			c.String("redis-host"),
			c.String("redis-password"),
			c.String("pidfile"),
			c.Int("concurrent"),
			c.Int("redis-db"),
			c.String("http-addr"),
			c.Bool("colorful-tty"),
			c.String("channel-default"),
			c.String("log-file"),
			c.Bool("debug"),
		)

		if os.Getuid() == 0 {
			fmt.Println(console.ColorfulText(
				console.TextYellow,
				"\n当前以root(%s)用户执行，使用root权限执行可能会造成严重的安全问题，建议使用非root用户执行\n",
			))
		}

		// 初始化日志输出
		// 指定日志文件时，使用日志文件输出，否则输出到标准输出
		if runtime.Config.LogFilename != "" {
			logFile, err := os.OpenFile(runtime.Config.LogFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				fmt.Printf("open log file %s failed: %v\n", runtime.Config.LogFilename, err)
				os.Exit(2)
			}
			defer logFile.Close()

			log.InitLogger(logFile, runtime.Config.DebugMode)
		} else {
			log.InitLogger(os.Stdout, runtime.Config.DebugMode)
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

		go func() {
			l, err := net.Listen("tcp", "0.0.0.0:10086")
			if err != nil {
				log.Error("创建socket连接失败: %v", err)
				return
			}
			defer l.Close()

			for {
				c, err := l.Accept()
				if err != nil {
					log.Error("socket连接失败: %v", err)
					return
				}

				go func(c net.Conn) {
					var wg sync.WaitGroup
					wg.Add(2)
					go func() {
						defer wg.Done()

						for {
							buffer := make([]byte, 1024)
							if _, err := c.Read(buffer); err != nil {
								break
							}

							fmt.Println(string(buffer))
							time.Sleep(2 * time.Second)
						}
					}()

					go func() {
						defer wg.Done()

						for {
							c.Write([]byte("Hello from server"))
							time.Sleep(2 * time.Second)
						}
					}()

					wg.Wait()
				}(c)
			}
		}()

		wg.Wait()
		log.Debug("all stoped.")

		return nil
	}

	app.Run(os.Args)
}
