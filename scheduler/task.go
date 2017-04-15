package scheduler

import (
	"sync"

	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	commander "github.com/mylxsw/coyotes/command"
	"github.com/mylxsw/coyotes/config"
)

// StartTaskRunner function start a taskRunner instance
func StartTaskRunner(channel *brokers.Channel) {
	runtime := config.GetRuntime()

	// 非任务模式自动返回
	if !runtime.Config.TaskMode {
		return
	}

	// 初始化channel关闭chan
	// 注意：这里最好设置为非堵塞方式，以加快进程退出速度
	channel.StopChan = make(chan struct{}, 1)
	defer close(channel.StopChan)

	// if _, err := client.Ping().Result(); err != nil {
	// 	log.Error("Failed connected to redis server: %s", err)
	// 	os.Exit(2)
	// }

	queue := broker.CreateTaskChannel(channel)
	defer queue.Close()

	queue.RegisterWorker(func(task brokers.Task, processID string) bool {
		status, _ := commander.CreateShellCommand(task, channel.OutputChan).Execute(processID)
		return status
	})

	var wg sync.WaitGroup
	wg.Add(channel.WorkerCount + 1)

	go queue.Listen(func() {
		wg.Done()
	})

	for index := 0; index < channel.WorkerCount; index++ {
		go queue.Work(func() {
			wg.Done()
		})
	}

	wg.Wait()
}
