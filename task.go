package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/elgs/gostrgen"
	"github.com/mylxsw/remote-tail/console"
	commander "github.com/mylxsw/task-runner/command"
	"github.com/mylxsw/task-runner/config"
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
				console.ColorfulText(console.TextRed, "["+output.ProcessID+"]"),
				console.ColorfulText(console.TextBlue, "$"),
				console.ColorfulText(console.TextGreen, output.Name),
				console.ColorfulText(console.TextMagenta, "->"),
				console.ColorfulText(console.TextYellow, output.Content),
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
