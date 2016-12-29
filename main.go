package main

import (
	"fmt"
	"os"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
	"github.com/mylxsw/task-runner/pidfile"
	"github.com/mylxsw/task-runner/scheduler"
	"github.com/mylxsw/task-runner/scheduler/channel"
	"github.com/mylxsw/task-runner/signal"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	server "github.com/mylxsw/task-runner/http"
)

func main() {

	runtime := config.GetRuntime()

	// 创建进程pid文件
	pid, err := pidfile.New(runtime.Config.PidFile)
	if err != nil {
		log.Error("failed to create pidfile: %v", err)
		os.Exit(2)
	}
	defer pid.Remove()

	if runtime.Config.ColorfulTTY {
		fmt.Println(console.ColorfulText(console.TextCyan, config.WelcomeMessage()))
	}

	log.Debug("redis addr: %s", runtime.Config.Redis.Addr)
	log.Debug("process ID: %d", os.Getpid())

	// 初始化所有channel，必须在初始化信号处理之前
	channel.InitChannels()
	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver()

	go server.StartHTTPServer()
	go func() {
		broker.TransferPrepareTask()
	}()
	scheduler.Schedule()

	<-runtime.StopHTTPServer
	log.Debug("all stoped.")
}
