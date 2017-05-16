package scheduler

import (
	"sync"

	"context"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
)

var newQueue = make(chan *brokers.Channel, 5)

// Schedule 函数用于开始任务调度器
func Schedule(ctx context.Context) {

	outputChan := make(chan brokers.Output, 20)
	defer close(outputChan)

	go func() {
		for output := range outputChan {
			log.Info(
				"[%s] %s -> %s",
				output.ProcessID,
				output.Task.TaskName,
				output.Content,
			)
		}
	}()

	var wg sync.WaitGroup

	runtime := config.GetRuntime()
	for index := range runtime.Channels {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()

			runtime.Channels[i].OutputChan = outputChan
			StartTaskRunner(ctx, runtime.Channels[i])
		}(index)
	}

	for {
		select {
		case <-ctx.Done():
			goto STOP
		case queueChannel := <-newQueue:
			wg.Add(1)
			go func() {
				defer wg.Done()

				queueChannel.OutputChan = outputChan
				StartTaskRunner(ctx, queueChannel)
			}()
		}
	}

STOP:

	wg.Wait()
	log.Debug("scheduler stoped")
}

// NewQueue 函数用于创建一个新的队列
func NewQueue(name string, distinct bool, workerCount int) error {
	queueChannel, err := NewChannel(name, distinct, workerCount)
	if err != nil {
		return err
	}
	newQueue <- queueChannel

	return nil
}
