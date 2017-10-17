package redis

import "fmt"

const (
	taskQueueKey       = "%s:tasks:queue"
	taskQueueExecKey   = taskQueueKey + ":exec"
	taskDistinctKey    = taskQueueKey + ":%s:dis"
	taskFailedQueueKey = taskQueueKey + ":failed"
	taskPrepareKey     = "task:prepare:queue"
	taskChannelsKey    = "task:channels"
	taskDelayQueueKey  = "task:delay:queue"
)

// TaskQueueKey 返回任务队列的KEY
func TaskQueueKey(channel string) string {
	return fmt.Sprintf(taskQueueKey, channel)
}

// TaskQueueExecKey 返回执行中的任务队列KEY
func TaskQueueExecKey(channel string) string {
	return fmt.Sprintf(taskQueueExecKey, channel)
}

// TaskQueueDistinctKey 返回任务的去重KEY
func TaskQueueDistinctKey(channel string, command string) string {
	return fmt.Sprintf(taskDistinctKey, channel, command)
}

// TaskFailedQueueKey 返回失败任务队列KEY
func TaskFailedQueueKey(channel string) string {
	return fmt.Sprintf(taskFailedQueueKey, channel)
}

// TaskPrepareQueueKey return the prepare key for queue
func TaskPrepareQueueKey() string {
	return taskPrepareKey
}

// TaskChannelsKey 返回所有channel信息存储的KEY
func TaskChannelsKey() string {
	return taskChannelsKey
}

// TaskDelayQueueKey 返回延时任务的key
func TaskDelayQueueKey() string {
	return taskDelayQueueKey
}
