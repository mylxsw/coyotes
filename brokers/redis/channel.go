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
	runtime       *config.Runtime
	client        *redis.Client
	channel       *brokers.Channel
	workerCount   int
	workerHandler func(command brokers.Task, processID string) bool
}

// CreateTaskChannel create a redis queue
func CreateTaskChannel(channel *brokers.Channel) *TaskChannel {
	client := createRedisClient()
	return &TaskChannel{
		client:      client,
		runtime:     config.GetRuntime(),
		channel:     channel,
		workerCount: 0,
	}
}

// Close task queue
func (queue *TaskChannel) Close() {
	queue.client.Close()
}

// Listen to the redis queue
func (queue *TaskChannel) Listen(dispose func ()) {
	defer dispose()
	// 非任务模式不启用队列监听
	if !queue.runtime.Config.TaskMode {
		return
	}

	log.Debug("listener %s started.", queue.channel.Name)
	defer log.Debug("listener %s stopped.", queue.channel.Name)

	for {
		select {
		case <-queue.channel.StopChan:
			close(queue.channel.Task)
			return
		default:
			res, err := queue.client.BRPop(2*time.Second, TaskQueueKey(queue.channel.Name)).Result()
			if err != nil {
				continue
			}

			task := decodeTask(res[1])

			queue.client.HSet(TaskQueueExecKey(queue.channel.Name), task.ID, res[1])
			queue.channel.Task <- task
		}

	}
}

// RegisterWorker 注册worker来消费队列
func (queue *TaskChannel) RegisterWorker(callback func(command brokers.Task, processID string) bool) {
	queue.workerHandler = callback
}

// Work 执行消费者worker
func (queue *TaskChannel) Work(dispose func ()) {
	defer dispose()

	queue.workerCount++
	processID := fmt.Sprintf("%s %d", queue.channel.Name, queue.workerCount)

	log.Debug("worker [%s] started.", console.ColorfulText(console.TextRed, processID))
	defer log.Debug("worker [%s] stopped.", console.ColorfulText(console.TextRed, processID))

	for {
		select {
		case task, ok := <-queue.channel.Task:
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

					distinctKey := TaskQueueDistinctKey(queue.channel.Name, task.TaskName)
					execKey := TaskQueueExecKey(queue.channel.Name)

					log.Debug(
						"[%s] clean %s %s ...",
						console.ColorfulText(console.TextRed, processID),
						distinctKey,
						execKey,
					)

					err := queue.client.Del(distinctKey).Err()
					if err != nil {
						log.Error(
							"[%s] delete key %s failed: %v",
							console.ColorfulText(console.TextRed, processID),
							distinctKey,
							err,
						)
					}

					err = queue.client.HDel(execKey, task.ID).Err()
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

				if queue.workerHandler(task, processID) {
					isSuccess = true
				}
			}(task)
		}
	}
}
