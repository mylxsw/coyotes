package scheduler

import (
	"sync"
	"time"

	"github.com/mylxsw/coyotes/backend"
	"github.com/mylxsw/coyotes/log"

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

	queue.RegisterWorker(func(task brokers.Task, processID string) (status bool, err error) {
		shellCommand := commander.CreateShellCommand(task, channel.OutputChan)
		task.ExecAt = time.Now()

		defer func() {

			if !status {
				task.FailedAt = time.Now()
			}

			// 每个任务都可以指定是否将执行结果写入后端存储
			if !task.WriteBackend {
				return
			}

			// 如果指定了后端存储，则需要记录执行结果
			if driver := backend.Default(); driver != nil {
				id, err := driver.Insert(task, backend.Result{
					IsSuccessful: status,
					Stdout:       shellCommand.StdoutString(),
					Stderr:       shellCommand.StderrString(),
				})
				if err != nil {
					log.Error("save result to backend storage failed: saved_id=%s, task_id=%s, process_id=%s, reason=%s", id, task.ID, processID, err.Error())
					return
				}

				log.Info("save result to backend storage: saved_id=%s, task_id=%s, process_id=%s", id, task.ID, processID)
			}

		}()

		status, err = shellCommand.Execute(processID)
		return
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
