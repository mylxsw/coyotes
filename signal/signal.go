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

	signalChan := make(chan os.Signal, 1)
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

				for _, channel := range runtime.Channels {
					channel.StopChan <- struct{}{}
				}

				runtime.StopScheduler <- struct{}{}
				runtime.StopHTTPServer <- struct{}{}
			}
		}
	}()

}
