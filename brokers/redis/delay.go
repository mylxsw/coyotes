package redis

import (
	"fmt"
	"time"

	"strconv"

	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/log"
	"gopkg.in/redis.v5"
	"context"
)

// AddDelayTask 新增一个延时任务到队列
func (manager *TaskManager) AddDelayTask(execTime time.Time, task brokers.Task) (id string, existence bool, err error) {
	if _, ok := manager.runtime.Channels[task.Channel]; !ok {
		return "", false, fmt.Errorf("task channel [%s] not exist", task.TaskName)
	}

	//  如果没有指定任务ID，则自动生成
	if task.ID == "" {
		task.ID = generateUUID()
	}

	log.Info("add delay task: [%s] %s -> %s", execTime.Format("2006-01-02 15:04:05"), task.TaskName, task.Channel)

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
		log.Debug("add delay task %s to channel %s: %s", tk.ID, tk.Channel, res)
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