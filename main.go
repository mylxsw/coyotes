package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/pidfile"
	"github.com/mylxsw/task-runner/signal"
)

var redisAddr = flag.String("host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
var redisPassword = flag.String("password", "", "redis连接密码")
var httpAddr = flag.String("http-addr", ":60001", "HTTP监控服务监听地址+端口")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")

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
	}

	// 创建进程pid文件
	pid, err := pidfile.New(*pidFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer pid.Remove()

	fmt.Println(console.ColorfulText(console.TextCyan, welcomeMessage()))

	log.Printf("The redis addr: %s", runtime.Redis.Addr)
	log.Printf("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver(runtime)

	go startHttpServer(runtime)
	startTaskRunner(runtime)
}
