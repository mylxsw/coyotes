package redis

import (
	"fmt"
	"time"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	redis "gopkg.in/redis.v5"
)

// TaskChannel is the queue object for redis broker
type TaskChannel struct {
	Runtime *config.Runtime
	Client  *redis.Client
	Channel *brokers.Channel
}

// CreateTaskChannel create a redis queue
func CreateTaskChannel(channel *brokers.Channel) *TaskChannel {
	client := createRedisClient()
	return &TaskChannel{
		Client:  client,
		Runtime: config.GetRuntime(),
		Channel: channel,
	}
}

// Close task queue
func (queue *TaskChannel) Close() {
	queue.Client.Close()
}

// Listen to the redis queue
func (queue *TaskChannel) Listen() {

	// 非任务模式不启用队列监听
	if !queue.Runtime.Config.TaskMode {
		return
	}

	log.Debug("listener %s started.", queue.Channel.Name)
	defer log.Debug("listener %s stopped.", queue.Channel.Name)

	for {
		select {
		case <-queue.Channel.StopChan:
			close(queue.Channel.Task)
			return
		default:
			res, err := queue.Client.BRPop(2*time.Second, TaskQueueKey(queue.Channel.Name)).Result()
			if err != nil {
				continue
			}

			task := decodeTask(res[1])

			queue.Client.HSet(TaskQueueExecKey(queue.Channel.Name), task.ID, res[1])
			queue.Channel.Task <- task
		}

	}
}

// Work function consuming the queue
func (queue *TaskChannel) Work(i int, callback func(command brokers.Task, processID string) bool) {
	processID := fmt.Sprintf("%s %d", queue.Channel.Name, i)

	log.Debug("worker [%s] started.", console.ColorfulText(console.TextRed, processID))
	defer log.Debug("worker [%s] stopped.", console.ColorfulText(console.TextRed, processID))

	for {
		select {
		case task, ok := <-queue.Channel.Task:
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

					distinctKey := TaskQueueDistinctKey(queue.Channel.Name, task.TaskName)
					execKey := TaskQueueExecKey(queue.Channel.Name)

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
