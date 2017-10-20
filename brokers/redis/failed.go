package redis

import (
	"fmt"
	"time"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/log"
)

// AddFailedTask 添加失败的任务到失败任务队列
func (manager *TaskManager) AddFailedTask(task brokers.Task) error {
	task.RetryCount++
	task.FailedAt = time.Now()

	log.Info("add failed task: id=%s, name=%s, channel=%s, retry_count=%d", task.ID, task.TaskName, task.Channel, task.RetryCount)

	return manager.client.HSet(TaskFailedQueueKey(task.Channel), task.ID, encodeTask(task)).Err()
}

// GetFailedTasks 返回channel中所有失败的任务
func (manager *TaskManager) GetFailedTasks(channel string) (map[string]brokers.Task, error) {
	res, err := manager.client.HGetAll(TaskFailedQueueKey(channel)).Result()
	if err != nil {
		log.Warning("query failed task failed: %v", err)
		return nil, err
	}

	tasks := make(map[string]brokers.Task, len(res))
	for key, val := range res {
		tasks[key] = decodeTask(val)
	}

	return tasks, nil
}

// GetFailedTask 查询失败的任务
func (manager *TaskManager) GetFailedTask(channel, taskID string) (brokers.Task, error) {
	taskFailedQueueKey := TaskFailedQueueKey(channel)
	val, err := manager.client.HGet(taskFailedQueueKey, taskID).Result()
	if err != nil {
		log.Warning("query failed task failed: %v", err)
		return brokers.Task{}, fmt.Errorf("query failed task failed: %v", err)
	}

	return decodeTask(val), nil
}

// RemoveFailedTask 从失败任务队列中移除任务
func (manager *TaskManager) RemoveFailedTask(channel string, taskID string) (brokers.Task, error) {
	task, err := manager.GetFailedTask(channel, taskID)
	if err == nil {
		manager.client.HDel(taskFailedQueueKey, taskID)
		log.Info("remove failed task: id=%s, name=%s, channel=%s, retry_count=%s", task.ID, task.TaskName, task.Channel, task.RetryCount)
	}

	return task, err
}

// RetryFailedTask 重试失败的任务
func (manager *TaskManager) RetryFailedTask(channel string, taskID string) error {
	tk, err := manager.RemoveFailedTask(channel, taskID)
	if err != nil {
		return fmt.Errorf("query task %s failed: %v", tk.ID, err)
	}

	_, existence, err := GetTaskManager().AddTask(tk)
	if err != nil {
		log.Error("add task %s failed: %v", tk.ID, err)
		return fmt.Errorf("add task %s failed: %v", tk.ID, err)
	}

	log.Info("retry failed task: id=%s, name=%s, channel=%s, retry_count=%s", tk.ID, tk.TaskName, tk.Channel, tk.RetryCount)

	res := "add"
	if existence {
		res = "repeat"
	}
	log.Debug("repush task %s to channel %s: %s", tk.ID, tk.Channel, res)

	return nil
}
