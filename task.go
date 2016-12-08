package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/elgs/gostrgen"
	commander "github.com/mylxsw/task-runner/command"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	redisQueue "github.com/mylxsw/task-runner/queue/redis"
	redis "gopkg.in/redis.v5"
)

func startTaskRunner(runtime *config.Runtime) {
	outputChan := make(chan commander.Output, 20)

	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Redis.Addr,
		Password: runtime.Redis.Password,
		DB:       runtime.Redis.DB,
	})
	defer client.Close()

	if _, err := client.Ping().Result(); err != nil {
		log.Fatalf("Error: %s", err)
	}

	queue := redisQueue.RedisQueue{
		Client:  client,
		Runtime: runtime,
	}

	go queue.Listen()

	go func() {
		for output := range outputChan {
			log.Printf(
				"%s%s %s %s %s",
				console.ColorfulText(runtime, console.TextRed, "["+output.ProcessID+"]"),
				console.ColorfulText(runtime, console.TextBlue, "$"),
				console.ColorfulText(runtime, console.TextGreen, output.Name),
				console.ColorfulText(runtime, console.TextMagenta, "->"),
				console.ColorfulText(runtime, console.TextYellow, output.Content),
			)
		}
	}()

	cmder := &commander.Command{
		Output: outputChan,
	}

	var wg sync.WaitGroup
	for index := 0; index < runtime.Concurrent; index++ {
		wg.Add(1)

		go func(i int, client *redis.Client) {
			defer wg.Done()

			queue.Work(i, func(cmd string) {
				routineName, _ := gostrgen.RandGen(10, gostrgen.Upper, "", "Ol")
				cmder.ExecuteTask(fmt.Sprintf("%s %d", routineName, index), cmd)
			})
		}(index, client)
	}

	wg.Wait()
}
