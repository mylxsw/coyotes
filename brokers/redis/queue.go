package redis

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mylxsw/task-runner/config"
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
		Runtime: runtime,
	}
}

// PushTask function push a task to queue
func PushTask(taskName string, channelName string, distinct bool) (interface{}, error) {
	client := createRedisClient()
	defer client.Close()

	log.Info("Push: %s -> %s", taskName, channelName)

	return pushToQueueCmd.Run(
		client,
		[]string{TaskQueueKey(channelName), TaskQueueDistinctKey(channelName, taskName)},
		taskName,
		distinct,
	).Result()
}

// QueryTask function query task queue status
func QueryTask(channel string) (tasks []Task, err error) {
	client := createRedisClient()
	defer client.Close()

	tasks = []Task{}
	vals, err := client.LRange(TaskQueueKey(channel), 0, client.LLen(TaskQueueKey(channel)).Val()).Result()
	if err != nil {
		return
	}

	for _, v := range vals {
		status, _ := strconv.Atoi(client.Get(TaskQueueDistinctKey(channel, v)).Val())
		tasks = append(tasks, Task{
			TaskName: v,
			// 0-去重key已过期，1-队列中
			Status: status,
		})
	}

	// 查询执行中的任务
	for _, v := range client.SMembers(TaskQueueExecKey(channel)).Val() {
		tasks = append(tasks, Task{
			TaskName: v,
			// 2-执行中
			Status: 2,
		})
	}

	return
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

	log.Debug("Queue Listener %s started.", channel.Name)
	defer log.Debug("Queue Listener %s stopped.", channel.Name)

	for {
		select {
		case <-queue.Runtime.Stoped:
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

	log.Debug("Task customer [%s] started.", processID)
	defer log.Debug("Task customer [%s] stopped.", processID)

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

					log.Debug("[%s] clean %s %s ...", processID, distinctKey, execKey)

					err := queue.Client.Del(distinctKey).Err()
					if err != nil {
						log.Error("Delete key %s failed: %v", distinctKey, err)
					}

					err = queue.Client.SRem(execKey, res).Err()
					if err != nil {
						log.Error("Remove key %s from %s: %v", res, execKey, err)
					}

					log.Info("[%s] time-consuming %v", processID, time.Since(startTime))
				}()

				callback(res, processID)
			}(res)
		}
	}
}
