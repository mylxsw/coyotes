package scheduler

import (
	"sync"

	"context"

	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	commander "github.com/mylxsw/coyotes/command"
	"github.com/mylxsw/coyotes/config"
)

// StartTaskRunner function start a taskRunner instance
func StartTaskRunner(ctx context.Context, channel *brokers.Channel) {
	runtime := config.GetRuntime()

	// 非任务模式自动返回
	if !runtime.Config.TaskMode {
		return
	}

	queue := broker.CreateTaskChannel(channel)
	defer queue.Close()

	queue.RegisterWorker(func(task brokers.Task, processID string) (bool, error) {
		status, err := commander.CreateShellCommand(task, channel.OutputChan).Execute(processID)
		return status, err
	})

	var wg sync.WaitGroup
	wg.Add(channel.WorkerCount + 1)

	go queue.Listen(ctx, func() {
		wg.Done()
	})

	for index := 0; index < channel.WorkerCount; index++ {
		go queue.Work(func() {
			wg.Done()
		})
	}

	wg.Wait()
}
