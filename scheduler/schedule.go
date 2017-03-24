package scheduler

import (
	"sync"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
	"github.com/mylxsw/task-runner/scheduler/channel"
	"github.com/mylxsw/task-runner/scheduler/task"
)

var runtime *config.Runtime
var newQueue chan *config.Channel

func init() {
	runtime = config.GetRuntime()
	newQueue = make(chan *config.Channel, 5)
}

// Schedule 函数用于开始任务调度器
func Schedule() {
	var wg sync.WaitGroup

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
