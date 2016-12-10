package redis

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
	"gopkg.in/redis.v5"
)

type RedisQueue struct {
	Runtime *config.Runtime
	Client  *redis.Client
}

type Task struct {
	TaskName string
	Status   int
}

const (
	taskQueueKey     = "%s:tasks:queue"
	taskQueueExecKey = taskQueueKey + ":exec"
	taskDistinctKey  = taskQueueKey + ":%s:dis"
)

var pushToQueueCmd = redis.NewScript(`
local element = redis.call("EXISTS", KEYS[2])
if element ~= 1 then
    redis.call("LPUSH", KEYS[1], ARGV[1])
	redis.call("SETEX", KEYS[2], 1800, '1')
end
return element
`)

// 任务队列key
func TaskQueueKey(channel string) string {
	return fmt.Sprintf(taskQueueKey, channel)
}

// 执行中的任务队列key
func TaskQueueExecKey(channel string) string {
	return fmt.Sprintf(taskQueueExecKey, channel)
}

// 任务队列去重key
func TaskQueueDistinctKey(channel string, command string) string {
	return fmt.Sprintf(taskDistinctKey, channel, command)
}

// 监听任务队列
func (queue *RedisQueue) Listen(channel *config.Channel) {
	// 非任务模式不启用队列监听
	if !queue.Runtime.Config.TaskMode {
		return
	}

	log.Info("Queue Listener started.")
	defer log.Info("Queue Listener stopped.")

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

// 任务队列消费
func (queue *RedisQueue) Work(i int, channel *config.Channel, callback func(command string, processID string)) {
	processID := fmt.Sprintf("%s %d", channel.Name, i)

	log.Info("Task customer [%s] started.", processID)
	defer log.Info("Task customer [%s] stopped.", processID)

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

					log.Info("[%s] clean %s %s ...", processID, distinctKey, execKey)

					err := queue.Client.Del(distinctKey).Err()
					if err != nil {
						log.Error("Error: %v", err)
					}

					err = queue.Client.SRem(execKey, res).Err()
					if err != nil {
						log.Error("Error: %v", err)
					}

					log.Info("[%s] time-consuming %v", processID, time.Since(startTime))
				}()

				callback(res, processID)
			}(res)
		}
	}
}

// 推送任务到任务队列
func PushTaskToQueue(client *redis.Client, taskName string, channelName string) (interface{}, error) {
	log.Info("Push: %s -> %s", taskName, channelName)
	return pushToQueueCmd.Run(client, []string{TaskQueueKey(channelName), TaskQueueDistinctKey(channelName, taskName)}, taskName).Result()
}

// 查询任务队列
func QueryTaskQueue(client *redis.Client, channel string) (tasks []Task, err error) {

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
