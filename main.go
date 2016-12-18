package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
	"github.com/mylxsw/task-runner/pidfile"
	"github.com/mylxsw/task-runner/signal"

	server "github.com/mylxsw/task-runner/http"
	task "github.com/mylxsw/task-runner/task"
)

func main() {

	runtime := config.GetRuntime()

	// 创建进程pid文件
	pid, err := pidfile.New(runtime.Config.PidFile)
	if err != nil {
		log.Error("Failed to create pidfile: %v", err)
		os.Exit(2)
	}
	defer pid.Remove()

	if runtime.Config.ColorfulTTY {
		fmt.Println(console.ColorfulText(console.TextCyan, config.WelcomeMessage()))
	}

	log.Debug("The redis addr: %s", runtime.Config.Redis.Addr)
	log.Debug("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver()

	go server.StartHTTPServer()

	var wg sync.WaitGroup

	for index := range runtime.Channels {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()
			task.StartTaskRunner(runtime.Channels[i])
		}(index)
	}

	wg.Wait()
}
