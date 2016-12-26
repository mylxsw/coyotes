package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/lunny/log"
)

// PrepareTask is the task that prepared to join queue
type PrepareTask struct {
	Name      string `json:"task"`
	Channel   string `json:"chan"`
	Timestamp int64  `json:"ts"`
}

// Task represent a task object
type Task struct {
	TaskName string `json:"task_name"`
	Channel  string `json:"channel"`
	Status   int    `json:"status"`
}

// PushTask function push a task to queue
func PushTask(taskName string, channelName string) (interface{}, error) {

	if _, ok := runtime.Channels[channelName]; !ok {
		return nil, fmt.Errorf("task channel not exist")
	}

	client := createRedisClient()
	defer client.Close()

	log.Info("push: %s -> %s", taskName, channelName)

	return pushToQueueCmd.Run(
		client,
		[]string{TaskQueueKey(channelName), TaskQueueDistinctKey(channelName, taskName)},
		taskName,
		runtime.Channels[channelName].Distinct,
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
			Status:  status,
			Channel: channel,
		})
	}

	// 查询执行中的任务
	for _, v := range client.SMembers(TaskQueueExecKey(channel)).Val() {
		tasks = append(tasks, Task{
			TaskName: v,
			// 2-执行中
			Status:  2,
			Channel: channel,
		})
	}

	return
}

// TransferPrepareTask 将prepare队列中的任务加入正式的任务队列
func TransferPrepareTask() {
	client := createRedisClient()
	defer client.Close()

	for {
		res, err := client.BRPop(2*time.Second, TaskPrepareQueueKey()).Result()
		if err != nil {
			continue
		}

		var task PrepareTask
		if err := json.Unmarshal([]byte(res[1]), &task); err == nil {
			PushTask(task.Name, task.Channel)
		}
	}
}
