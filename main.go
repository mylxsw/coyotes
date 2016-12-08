package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
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

func main() {

	flag.Parse()

	runtime := &config.Runtime{
		PidFile:    *pidFile,
		Concurrent: *concurrent,
		Redis: config.RedisConfig{
			Addr:     *redisAddr,
			Password: *redisPassword,
		},
		Http: config.HttpConfig{
			ListenAddr: *httpAddr,
		},
		StopRunning:     false,
		StopRunningChan: make(chan struct{}, *concurrent),
		Command:         make(chan string, *concurrent),
		TaskMode:        *taskMode,
		ColorfulTTY:     *colorfulTTY,
	}

	// 创建进程pid文件
	pid, err := pidfile.New(*pidFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer pid.Remove()

	fmt.Println(console.ColorfulText(runtime, console.TextCyan, welcomeMessage(runtime)))

	log.Printf("The redis addr: %s", runtime.Redis.Addr)
	log.Printf("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver(runtime)

	go startHttpServer(runtime)
	startTaskRunner(runtime)
}
