package redis

import (
	"fmt"
	"time"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
	redis "gopkg.in/redis.v5"
)

// Queue is the queue object for redis broker
type Queue struct {
	Runtime *config.Runtime
	Client  *redis.Client
}

// Create a redis queue
func Create() *Queue {
	client := createRedisClient()
	return &Queue{
		Client:  client,
		Runtime: config.GetRuntime(),
	}
}

// Close task queue
func (queue *Queue) Close() {
	queue.Client.Close()
}

// Listen to the redis queue
func (queue *Queue) Listen(channel *config.Channel) {

	// 非任务模式不启用队列监听
	if !queue.Runtime.Config.TaskMode {
		return
	}

	log.Debug("queue listener %s started.", channel.Name)
	defer log.Debug("queue listener %s stopped.", channel.Name)

	for {
		select {
		case <-channel.StopChan:
			close(channel.Command)
			return
		default:
			res, err := queue.Client.BRPop(2*time.Second, TaskQueueKey(channel.Name)).Result()
			if err != nil {
				continue
			}

			queue.Client.SAdd(TaskQueueExecKey(channel.Name), res[1])
			channel.Command <- res[1]
		}

	}
}

// Work function consuming the queue
func (queue *Queue) Work(i int, channel *config.Channel, callback func(command string, processID string)) {
	processID := fmt.Sprintf("%s %d", channel.Name, i)

	log.Debug("task customer [%s] started.", console.ColorfulText(console.TextRed, processID))
	defer log.Debug("task customer [%s] stopped.", console.ColorfulText(console.TextRed, processID))

	for {
		select {
		case res, ok := <-channel.Command:
			if !ok {
				return
			}

			func(res string) {

				startTime := time.Now()

				// 删除用于去重的缓存key
				defer func() {
					distinctKey := TaskQueueDistinctKey(channel.Name, res)
					execKey := TaskQueueExecKey(channel.Name)

					log.Debug(
						"[%s] clean %s %s ...",
						console.ColorfulText(console.TextRed, processID),
						distinctKey,
						execKey,
					)

					err := queue.Client.Del(distinctKey).Err()
					if err != nil {
						log.Error(
							"[%s] delete key %s failed: %v",
							console.ColorfulText(console.TextRed, processID),
							distinctKey,
							err,
						)
					}

					err = queue.Client.SRem(execKey, res).Err()
					if err != nil {
						log.Error(
							"[%s] remove key %s from %s: %v",
							console.ColorfulText(console.TextRed, processID),
							res,
							execKey,
							err,
						)
					}

					log.Info(
						"[%s] task [%s] time-consuming %v",
						console.ColorfulText(console.TextRed, processID),
						console.ColorfulText(console.TextGreen, res),
						time.Since(startTime),
					)
				}()

				callback(res, processID)
			}(res)
		}
	}
}
