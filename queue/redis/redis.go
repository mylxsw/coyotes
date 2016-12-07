package redis

import (
	"fmt"
	"log"
	"time"

	"github.com/mylxsw/task-runner/config"
	"gopkg.in/redis.v5"
)

type RedisQueue struct {
	Runtime *config.Runtime
	Client  *redis.Client
}

func (queue *RedisQueue) Listen() {

	log.Print("Queue Listener started.")

	for {
		if queue.Runtime.StopRunning {
			return
		}
		res, err := queue.Client.BRPop(5*time.Second, "tasks:async:queue").Result()
		if err != nil {
			continue
		}

		queue.Runtime.Command <- res[1]
	}

	log.Print("Queue Listener stopped.")
}

func (queue *RedisQueue) Work(i int, callback func(command string)) {
	defer func() {
		log.Printf("Task customer [%d] stopped.", i)
	}()

	log.Printf("Task customer [%d] started.", i)

	for {
		// worker exit
		if queue.Runtime.StopRunning {
			return
		}

		select {
		case res := <-queue.Runtime.Command:
			func(res string) {
				// 命令执行完毕才删除去重的key
				// TODO 这个set是兼容已有方案用的，下次更新的时候去掉即可
				defer queue.Client.SRem("tasks:async:queue:distinct", res)
				// 删除用于去重的缓存key
				defer queue.Client.Del(fmt.Sprintf("tasks:distinct:%s", res))

				callback(res)
			}(res)
		case <-queue.Runtime.StopRunningChan:
			return
		}
	}
}
