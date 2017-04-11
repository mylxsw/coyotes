package redis

import (
	"fmt"
	"time"

	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	redis "gopkg.in/redis.v5"
	"github.com/mylxsw/coyotes/brokers"
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
func (queue *Queue) Listen(channel *brokers.Channel) {

	// 非任务模式不启用队列监听
	if !queue.Runtime.Config.TaskMode {
		return
	}

	log.Debug("queue listener %s started.", channel.Name)
	defer log.Debug("queue listener %s stopped.", channel.Name)

	for {
		select {
		case <-channel.StopChan:
			close(channel.Task)
			return
		default:
			res, err := queue.Client.BRPop(2*time.Second, TaskQueueKey(channel.Name)).Result()
			if err != nil {
				continue
			}

			task := decodeTask(res[1])

			queue.Client.HSet(TaskQueueExecKey(channel.Name), task.ID, res[1])
			channel.Task <- task
		}

	}
}

// Work function consuming the queue
func (queue *Queue) Work(i int, channel *brokers.Channel, callback func(command brokers.Task, processID string) bool) {
	processID := fmt.Sprintf("%s %d", channel.Name, i)

	log.Debug("task customer [%s] started.", console.ColorfulText(console.TextRed, processID))
	defer log.Debug("task customer [%s] stopped.", console.ColorfulText(console.TextRed, processID))

	for {
		select {
		case task, ok := <-channel.Task:
			if !ok {
				return
			}

			func(task brokers.Task) {

				startTime := time.Now()
				// 执行结果，默认为false，失败
				isSuccess := false

				// 删除用于去重的缓存key
				defer func() {

					// 统计任务执行结果
					if isSuccess {
						config.IncrSuccTaskCount()
					} else {
						config.IncrFailTaskCount()
					}

					distinctKey := TaskQueueDistinctKey(channel.Name, task.TaskName)
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

					err = queue.Client.HDel(execKey, task.ID).Err()
					if err != nil {
						log.Error(
							"[%s] remove key %s from %s: %v",
							console.ColorfulText(console.TextRed, processID),
							task.TaskName,
							execKey,
							err,
						)
					}

					log.Info(
						"[%s] task [%s] time-consuming %v",
						console.ColorfulText(console.TextRed, processID),
						console.ColorfulText(console.TextGreen, task.TaskName),
						time.Since(startTime),
					)
				}()

				if callback(task, processID) {
					isSuccess = true
				}
			}(task)
		}
	}
}
