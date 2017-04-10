package scheduler

import (
	"sync"

	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/scheduler/channel"
	"github.com/mylxsw/coyotes/scheduler/task"
)

var newQueue = make(chan *config.Channel, 5)

// Schedule 函数用于开始任务调度器
func Schedule() {
	var wg sync.WaitGroup

	runtime := config.GetRuntime()
	for index := range runtime.Channels {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()
			task.StartTaskRunner(runtime.Channels[i])
		}(index)
	}

	for {
		select {
		case <-runtime.StopScheduler:
			goto STOP
		case queueChannel := <-newQueue:
			wg.Add(1)
			go func() {
				defer wg.Done()
				task.StartTaskRunner(queueChannel)
			}()
		}
	}

STOP:

	wg.Wait()
	log.Debug("scheduler stoped.")
}

// NewQueue 函数用于创建一个新的队列
func NewQueue(name string, distinct bool, workerCount int) error {
	queueChannel, err := channel.NewChannel(name, distinct, workerCount)
	if err != nil {
		return err
	}
	newQueue <- queueChannel

	return nil
}
