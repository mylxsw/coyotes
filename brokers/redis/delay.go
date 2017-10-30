package redis

import (
	"fmt"
	"time"

	"strconv"

	"context"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/log"
	"gopkg.in/redis.v5"
)

// AddDelayTask 新增一个延时任务到队列
func (manager *TaskManager) AddDelayTask(execTime time.Time, task brokers.Task) (id string, existence bool, err error) {
	// 延迟任务自动加入到延迟任务队列
	task.Channel = "delayed"

	if _, ok := manager.runtime.Channels[task.Channel]; !ok {
		return "", false, fmt.Errorf("task channel [%s] not exist", task.TaskName)
	}

	//  如果没有指定任务ID，则自动生成
	if task.ID == "" {
		task.ID = generateUUID()
	}

	task.ExecAt = execTime

	log.Info("add delay task: id=%s, name=%s, channel=%s, exec_at=%s", task.ID, task.TaskName, task.Channel, execTime.Format("2006-01-02 15:04:05"))

	score, _ := strconv.ParseFloat(execTime.Format("20060102150405"), 64)
	val, err := manager.client.ZAdd(TaskDelayQueueKey(), redis.Z{
		Score:  score,
		Member: encodeTask(task),
	}).Result()

	if err != nil {
		return "", false, err
	}

	return task.ID, val != 1, nil
}

// GetDelayTasks 获取所有延迟任务
func (manager *TaskManager) GetDelayTasks() (map[string]brokers.Task, error) {
	vals, err := manager.client.ZRange(TaskDelayQueueKey(), 0, -1).Result()
	if err != nil {
		log.Warning("query delay tasks failed:%v", err)
		return nil, err
	}

	tasks := make(map[string]brokers.Task, len(vals))
	for _, val := range vals {
		tk := decodeTask(val)
		tasks[tk.ID] = tk
	}

	return tasks, nil
}

// GetDelayTask Get specified delay task
func (manager *TaskManager) GetDelayTask(taskID string) (brokers.Task, error) {
	// TODO 临时实现，后面考虑新的数据存储结构和后端
	vals, err := manager.client.ZRange(TaskDelayQueueKey(), 0, -1).Result()
	if err != nil {
		log.Warning("query delay tasks failed:%v", err)
		return brokers.Task{}, err
	}

	for _, val := range vals {
		tk := decodeTask(val)
		if tk.ID == taskID {
			return tk, nil
		}
	}

	return brokers.Task{}, fmt.Errorf("Can't find the specified delay task")
}

// RemoveDelayTask Remove a delay task from queue
func (manager *TaskManager) RemoveDelayTask(taskID string) (brokers.Task, error) {
	// TODO 临时实现，后面考虑新的数据存储结构和后端
	vals, err := manager.client.ZRange(TaskDelayQueueKey(), 0, -1).Result()
	if err != nil {
		log.Warning("query delay tasks failed:%v", err)
		return brokers.Task{}, err
	}

	for _, val := range vals {
		tk := decodeTask(val)
		if tk.ID == taskID {
			if err := manager.client.ZRem(TaskDelayQueueKey(), val).Err(); err != nil {
				return brokers.Task{}, err
			}

			log.Info("remove delay task: id=%s, name=%s", tk.ID, tk.TaskName)

			return tk, nil
		}
	}

	return brokers.Task{}, fmt.Errorf("Can't find the specified delay task")
}

// MigrateDelayTask 迁移延时任务到执行队列
func (manager *TaskManager) MigrateDelayTask() {
	res, err := popDelayTasks.Run(
		manager.client,
		[]string{TaskDelayQueueKey()},
		time.Now().Format("20060102150405"),
	).Result()

	if err != nil {
		log.Warning("query delay task failed: %v", err)
		return
	}

	for _, t := range res.([]interface{}) {

		tk := decodeTask(t.(string))
		_, existence, err := GetTaskManager().AddTask(tk)

		if err != nil {
			log.Error("add task %s failed: %v", tk.ID, err)
			continue
		}

		res := "add"
		if existence {
			res = "repeat"
		}
		log.Debug("%s delay task to execute queue: id=%s, name=%s, channel=%s", res, tk.ID, tk.TaskName, tk.Channel)
	}
}

// StartDelayTaskLifeCycle 启动延时任务迁移
func StartDelayTaskLifeCycle(ctx context.Context) {
	for {
		time.Sleep(time.Second)
		select {
		case <-ctx.Done():
			return
		default:
			GetTaskManager().MigrateDelayTask()
		}
	}
}
