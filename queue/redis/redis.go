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
-- KEYS[1]=队列key
-- KEYS[2]=去重key
-- ARGV[1]=命令
-- ARGV[2]=是否启用去重

-- redis.log(redis.LOG_NOTICE, type(ARGV[2]) .. " : " .. ARGV[2])

-- 如果不启用去重复功能，则直接push到任务队列
if ARGV[2] == '0' then 
	redis.call("LPUSH", KEYS[1], ARGV[1])
	return 1
end

-- 不启用去重复功能，先判断是否存在去重key，存在则不添加队列
-- 不存在则添加队列并设置去重key，有效期1800s
local element = redis.call("EXISTS", KEYS[2])
if element ~= 1 then
	redis.call("LPUSH", KEYS[1], ARGV[1])
	redis.call("SETEX", KEYS[2], 1800, '1')
	
	return 1
end

return 0
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

// 任务队列消费
func (queue *RedisQueue) Work(i int, channel *config.Channel, callback func(command string, processID string)) {
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

// 推送任务到任务队列
func PushTaskToQueue(client *redis.Client, taskName string, channelName string, distinct bool) (interface{}, error) {
	log.Info("Push: %s -> %s", taskName, channelName)
	return pushToQueueCmd.Run(
		client,
		[]string{TaskQueueKey(channelName), TaskQueueDistinctKey(channelName, taskName)},
		taskName,
		distinct,
	).Result()
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
