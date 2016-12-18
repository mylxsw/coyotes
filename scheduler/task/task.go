package task

import (
	"sync"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	commander "github.com/mylxsw/task-runner/command"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
)

var runtime *config.Runtime

func init() {
	runtime = config.GetRuntime()
}

// StartTaskRunner function start a taskRunner instance
func StartTaskRunner(channel *config.Channel) {
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
				"%s%s %s %s %s",
				console.ColorfulText(console.TextRed, "["+output.ProcessID+"]"),
				console.ColorfulText(console.TextBlue, "$"),
				console.ColorfulText(console.TextGreen, output.Name),
				console.ColorfulText(console.TextMagenta, "->"),
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
