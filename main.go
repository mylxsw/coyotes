package main

import (
	"gopkg.in/redis.v4"
	"sync"
	"fmt"
	"os/exec"
	"strings"
	"io"
	"bufio"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/mylxsw/task-runner/pidfile"
	"flag"
)

var redisAddr = flag.String("redis-host", "127.0.0.1:6379", "redis连接地址")
var pidFile = flag.String("pidfile", "/tmp/task-runner.pid", "pid文件路径")
var concurrent = flag.Int("concurrent", 5, "并发执行线程数")

var stopRunning bool = false
var command chan string
var outputs chan string = make(chan string, 20)

func main() {

	flag.Parse()

	command = make(chan string, *concurrent)

	// 创建进程pid文件
	pid, err := pidfile.New(*pidFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer pid.Remove()

	log.Printf("The process ID: %d", os.Getpid())

	// 信号处理程序，接收退出信号，平滑退出进程
	initSignalReceiver()

	client := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
		Password: "",
		DB:0,
	})
	defer client.Close()

	initQueueListener(client)
	initOutput()

	var wg sync.WaitGroup
	for i := 0; i < *concurrent; i ++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			worker(i)
		}(i)
	}

	wg.Wait()
}

// worker，消费队列中的命令
func worker(i int) {
	defer func() {
		log.Printf("Task customer [%d] stopped.", i)
	}()

	log.Printf("Task customer [%d] started.", i)

	for {
		// worker exit
		if stopRunning {
			return
		}

		select {
		case res := <-command:
			params := strings.Split(res, " ")
			executeTask(outputs, params)
		}
	}
}

// 执行命令，绑定输出
func executeTask(output chan string, params []string) error {
	cmd := exec.Command(params[0], params[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		bindOutput(params[0], output, &stdout)
	}()
	go func() {
		defer wg.Done()
		bindOutput(params[0], output, &stderr)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// 初始化信号接受处理程序
func initSignalReceiver() {
	signalChan := make(chan os.Signal)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	go func() {
		for {
			sig := <-signalChan
			switch sig {
			case syscall.SIGUSR2, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL:
				stopRunning = true
				close(command)
				log.Print("Received exit signal.")
			}
		}
	}()

}

// 初始化队列监听，监听队列里面的新到达数据
func initQueueListener(client *redis.Client) {
	go func() {

		log.Print("Queue Listener started.")

		for {
			if stopRunning {
				return
			}
			res, err := client.BRPop(5 * time.Second, "tasks:async:queue").Result()
			if err != nil {
				continue
			}

			command <- res[1]
		}

		log.Print("Queue Listener stopped.")
	}()
}

// 初始化输出
func initOutput() {
	go func() {
		for output := range outputs {
			log.Printf(
				"-> %s",
				output,
			)
		}
	}()
}

// 绑定标准输入、输出到输出channel
func bindOutput(name string, output chan string, input *io.ReadCloser) error {
	reader := bufio.NewReader(*input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			if err != io.EOF {
				return fmt.Errorf("命令执行失败: %s", err)
			}
			break
		}

		output <- fmt.Sprintf("%s -> %s", name, line)
	}

	return nil
}