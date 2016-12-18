package task

import (
	"os"
	"sync"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	commander "github.com/mylxsw/task-runner/command"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
	redis "gopkg.in/redis.v5"
)

// StartTaskRunner function start a taskRunner instance
func StartTaskRunner(runtime *config.Runtime, channel *config.Channel) {
	outputChan := make(chan commander.Output, 20)
	defer close(outputChan)

	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Config.Redis.Addr,
		Password: runtime.Config.Redis.Password,
		DB:       runtime.Config.Redis.DB,
	})
	defer client.Close()

	if _, err := client.Ping().Result(); err != nil {
		log.Error("Failed connected to redis server: %s", err)
		os.Exit(2)
	}

	queue := broker.RedisQueue{
		Client:  client,
		Runtime: runtime,
	}

	go func() {
		for output := range outputChan {
			log.Info(
				"%s%s %s %s %s",
				console.ColorfulText(runtime, console.TextRed, "["+output.ProcessID+"]"),
				console.ColorfulText(runtime, console.TextBlue, "$"),
				console.ColorfulText(runtime, console.TextGreen, output.Name),
				console.ColorfulText(runtime, console.TextMagenta, "->"),
				console.ColorfulText(runtime, console.TextYellow, output.Content),
			)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(runtime.Config.Concurrent + 1)

	go func() {
		defer wg.Done()
		queue.Listen(channel)
	}()

	for index := 0; index < runtime.Config.Concurrent; index++ {
		go func(i int, client *redis.Client, channel *config.Channel) {
			defer wg.Done()

			queue.Work(i, channel, func(cmd string, processID string) {
				cmder := &commander.Command{
					Output: outputChan,
				}
				cmder.ExecuteTask(processID, cmd)
			})
		}(index, client, channel)
	}

	wg.Wait()
}
