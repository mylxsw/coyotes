package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
)

// InitSignalReceiver 初始化信号接受处理程序
func InitSignalReceiver() {

	runtime := config.GetRuntime()

	// 用于向所有channel发送程序退出信号
	// TODO 新增channel后如何更新该值？
	runtime.Stoped = make(chan struct{}, len(runtime.Channels))
	runtime.StopHTTPServer = make(chan struct{})
	runtime.StopScheduler = make(chan struct{})

	signalChan := make(chan os.Signal)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	go func() {

		for {
			sig := <-signalChan
			switch sig {
			case syscall.SIGUSR2, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL:
				log.Debug("Received exit signal, Waiting for exit...")

				for i := 0; i < len(runtime.Channels); i++ {
					runtime.Stoped <- struct{}{}
				}

				runtime.StopScheduler <- struct{}{}
				runtime.StopHTTPServer <- struct{}{}
			}
		}
	}()

}
