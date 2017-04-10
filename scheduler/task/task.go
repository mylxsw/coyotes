package task

import (
	"sync"

	broker "github.com/mylxsw/coyotes/brokers/redis"
	commander "github.com/mylxsw/coyotes/command"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
)

// StartTaskRunner function start a taskRunner instance
func StartTaskRunner(channel *config.Channel) {
	runtime := config.GetRuntime()

	// 非任务模式自动返回
	if !runtime.Config.TaskMode {
		return
	}

	// 初始化channel关闭chan
	// 注意：这里最好设置为非堵塞方式，以加快进程退出速度
	channel.StopChan = make(chan struct{}, 1)
	defer close(channel.StopChan)

	outputChan := make(chan commander.Output, 20)
	defer close(outputChan)

	// if _, err := client.Ping().Result(); err != nil {
	// 	log.Error("Failed connected to redis server: %s", err)
	// 	os.Exit(2)
	// }

	queue := broker.Create()
	defer queue.Close()

	go func() {
		for output := range outputChan {
			log.Info(
				"[%s] %s -> %s",
				console.ColorfulText(console.TextRed, output.ProcessID),
				console.ColorfulText(console.TextGreen, output.Name),
				console.ColorfulText(console.TextYellow, output.Content),
			)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(channel.WorkerCount + 1)

	go func() {
		defer wg.Done()
		queue.Listen(channel)
	}()

	for index := 0; index < channel.WorkerCount; index++ {
		go func(i int) {
			defer wg.Done()

			queue.Work(i, channel, func(cmd string, processID string) {
				cmder := &commander.Command{
					Output: outputChan,
				}
				cmder.ExecuteTask(processID, cmd)
			})
		}(index)
	}

	wg.Wait()
}
