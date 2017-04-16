package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mylxsw/coyotes/log"
	"context"
)

// InitSignalReceiver 初始化信号接受处理程序
func InitSignalReceiver(ctx context.Context, cancel context.CancelFunc) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	go func() {

		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGUSR2, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL:
				log.Debug("waiting for exit...")
				cancel()
			}
		}
	}()

}
