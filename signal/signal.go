package signal

import (
	"os"
	"os/signal"
	"syscall"
	"log"
)

// 初始化信号接受处理程序
func InitSignalReceiver(concurrent int, stopRunning *bool, stopRunningChan *chan struct{}) {
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
				*stopRunning = true
				//close(command)
				for i := 0; i < concurrent; i++ {
					*stopRunningChan <- struct{}{}
				}
				log.Print("Received exit signal.")
			}
		}
	}()

}
