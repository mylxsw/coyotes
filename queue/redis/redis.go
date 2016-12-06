package redis

import (
	"gopkg.in/redis.v4"
	"log"
	"time"
	"fmt"
)

type RedisQueue struct {
	StopRunning bool
	StopRunningChan chan struct{}
	Command chan string
	Client *redis.Client
}

func (queue *RedisQueue) Listen() {

	log.Print("Queue Listener started.")

	for {
		if queue.StopRunning {
			return
		}
		res, err := queue.Client.BRPop(5*time.Second, "tasks:async:queue").Result()
		if err != nil {
			continue
		}

		queue.Command <- res[1]
	}

	log.Print("Queue Listener stopped.")
}

func (queue *RedisQueue) Work(i int, callback func (command string)) {
	defer func() {
		log.Printf("Task customer [%d] stopped.", i)
	}()

	log.Printf("Task customer [%d] started.", i)

	for {
		// worker exit
		if queue.StopRunning {
			return
		}

		select {
		case res := <-queue.Command:
			func(res string) {
				// 命令执行完毕才删除去重的key
				// TODO 这个set是兼容已有方案用的，下次更新的时候去掉即可
				defer queue.Client.SRem("tasks:async:queue:distinct", res)
				// 删除用于去重的缓存key
				defer queue.Client.Del(fmt.Sprintf("tasks:distinct:%s", res))

				callback(res)
			}(res)
		case <-queue.StopRunningChan:
			return
		}
	}
}

