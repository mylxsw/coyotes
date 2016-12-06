package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/mylxsw/task-runner/pidfile"
	"gopkg.in/redis.v4"
	"github.com/mylxsw/task-runner/signal"
	redisQueue "github.com/mylxsw/task-runner/queue/redis"
	commander "github.com/mylxsw/task-runner/command"
	"github.com/mylxsw/remote-tail/console"
	"github.com/elgs/gostrgen"
	"fmt"
)

var redisAddr = flag.String("host", "127.0.0.1:6379", "redis连接地址，必须指定端口")
var redisPassword = flag.String("password", "", "redis连接密码")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")

var stopRunning bool = false
var stopRunningChan chan struct{}
var command chan string
var outputs chan commander.Output = make(chan commander.Output, 20)

func main() {

	flag.Parse()

	command = make(chan string, *concurrent)
	stopRunningChan = make(chan struct{}, *concurrent)

	// 创建进程pid文件
	pid, err := pidfile.New(*pidFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer pid.Remove()

	fmt.Println(console.ColorfulText(console.TextCyan, welcomeMessage()))

	log.Printf("The redis addr: %s", *redisAddr)
	log.Printf("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	signal.InitSignalReceiver(*concurrent, &stopRunning, &stopRunningChan)

	client := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       0,
	})
	defer client.Close()

	if _, err := client.Echo("Hello").Result(); err != nil {
		log.Fatalf("Error: %s", err)
	}

	queue := redisQueue.RedisQueue{
		Client:client,
		StopRunning: stopRunning,
		StopRunningChan: stopRunningChan,
		Command:command,
	}

	go queue.Listen()

	go func() {
		for output := range outputs {
			log.Printf(
				"%s%s %s %s %s",
				console.ColorfulText(console.TextRed, "[" + output.ProcessID + "]"),
				console.ColorfulText(console.TextBlue, "$"),
				console.ColorfulText(console.TextGreen, output.Name),
				console.ColorfulText(console.TextMagenta, "->"),
				console.ColorfulText(console.TextYellow, output.Content),
			)
		}
	}()

	cmder := &commander.Command{
		Output:outputs,
	}

	var wg sync.WaitGroup
	for index := 0; index < *concurrent; index++ {
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

