package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
)

// 初始化信号接受处理程序
func InitSignalReceiver(runtime *config.Runtime) {
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
				log.Info("Received exit signal, Waiting for exit...")

				for i := 0; i < len(runtime.Channels); i++ {
					runtime.Stoped <- struct{}{}
				}
			}
		}
	}()

}
