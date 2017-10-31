package redis

import (
	"fmt"
	"sync"
	"time"

	"context"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
	redis "gopkg.in/redis.v5"
)

// TaskChannel is the queue object for redis broker
type TaskChannel struct {
	runtime        *config.Runtime
	client         *redis.Client
	channel        *brokers.Channel
	workerCount    int
	workerCountMux sync.Mutex
	workerHandler  func(command brokers.Task, processID string) (bool, error)
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
func (queue *TaskChannel) Listen(ctx context.Context, dispose func()) {
	defer dispose()
	// 非任务模式不启用队列监听
	if !queue.runtime.Config.TaskMode {
		return
	}

	log.Debug("listener %s started.", queue.channel.Name)
	defer log.Debug("listener %s stopped.", queue.channel.Name)

	for {
		select {
		case <-ctx.Done():
			close(queue.channel.Task)
			return
		default:
			// 从队列中取出一个任务
			// 1. 从TaskQueueKey中取出一个任务
			// 2. 在TaskQueueExecKey这个Hash表中添加该任务，标识该任务正在执行
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
func (queue *TaskChannel) RegisterWorker(callback func(command brokers.Task, processID string) (bool, error)) {
	queue.workerHandler = callback
}

// NewWorkerProcessID 为worker分配ID
func (queue *TaskChannel) NewWorkerProcessID() string {
	queue.workerCountMux.Lock()
	defer queue.workerCountMux.Unlock()

	queue.workerCount++

	return fmt.Sprintf("%s %d", queue.channel.Name, queue.workerCount)
}

// Work 执行消费者worker
func (queue *TaskChannel) Work(dispose func()) {
	defer dispose()

	processID := queue.NewWorkerProcessID()

	log.Debug("worker [%s] started.", processID)
	defer log.Debug("worker [%s] stopped.", processID)

	for {
		select {
		case task, ok := <-queue.channel.Task:
			if !ok {
				return
			}

			func(task brokers.Task) {

				startTime := time.Now()

				// 执行结果，默认为false，失败
				var taskErr error
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
						processID,
						distinctKey,
						execKey,
					)

					// 任务执行完成后处理
					// 1. 从TaskQueueDistinctKey中去除任务去重Key
					// 2. 从TaskQueueExecKey中移除当前任务的ID，标识该任务已经执行结束

					err := queue.client.Del(distinctKey).Err()
					if err != nil {
						log.Error(
							"[%s] delete key %s failed: %v",
							processID,
							distinctKey,
							err,
						)
					}

					err = queue.client.HDel(execKey, task.ID).Err()
					if err != nil {
						log.Error(
							"[%s] remove key %s from %s: %v",
							processID,
							task.TaskName,
							execKey,
							err,
						)
					}

					// 如果任务执行失败，则需要将其重新加入到失败任务队列
					if !isSuccess {
						GetTaskManager().AddFailedTask(task)
					}

					log.Info(
						"[%s] task [%s] time-consuming %v",
						processID,
						task.TaskName,
						time.Since(startTime),
					)
				}()

				// 执行取出的任务
				if isSuccess, taskErr = queue.workerHandler(task, processID); taskErr != nil {
					log.Error("[%s] task [%s] execute failed: %v", processID, task.TaskName, taskErr)
				}
			}(task)
		}
	}
}
