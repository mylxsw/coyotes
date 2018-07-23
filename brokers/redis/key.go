package redis

import (
	"fmt"

	"github.com/mylxsw/coyotes/config"
)

const (
	taskQueueKey       = "%s:%s:tasks:queue"
	taskQueueExecKey   = taskQueueKey + ":exec"
	taskDistinctKey    = taskQueueKey + ":%s:dis"
	taskFailedQueueKey = taskQueueKey + ":failed"
	taskPrepareKey     = "%s:task:prepare:queue"
	taskChannelsKey    = "%s:task:channels"
	taskDelayQueueKey  = "%s:task:delay:queue"
)

// TaskQueueKey 返回任务队列的KEY
func TaskQueueKey(channel string) string {
	return fmt.Sprintf(taskQueueKey, config.GetRuntime().Config.BizName, channel)
}

// TaskQueueExecKey 返回执行中的任务队列KEY
func TaskQueueExecKey(channel string) string {
	return fmt.Sprintf(taskQueueExecKey, config.GetRuntime().Config.BizName, channel)
}

// TaskQueueDistinctKey 返回任务的去重KEY
func TaskQueueDistinctKey(channel string, command string) string {
	return fmt.Sprintf(taskDistinctKey, config.GetRuntime().Config.BizName, channel, command)
}

// TaskFailedQueueKey 返回失败任务队列KEY
func TaskFailedQueueKey(channel string) string {
	return fmt.Sprintf(taskFailedQueueKey, config.GetRuntime().Config.BizName, channel)
}

// TaskPrepareQueueKey return the prepare key for queue
func TaskPrepareQueueKey() string {
	return fmt.Sprintf(taskPrepareKey, config.GetRuntime().Config.BizName)
}

// TaskChannelsKey 返回所有channel信息存储的KEY
func TaskChannelsKey() string {
	return fmt.Sprintf(taskChannelsKey, config.GetRuntime().Config.BizName)
}

// TaskDelayQueueKey 返回延时任务的key
func TaskDelayQueueKey() string {
	return fmt.Sprintf(taskDelayQueueKey, config.GetRuntime().Config.BizName)
}
