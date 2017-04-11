package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
	"github.com/docker/distribution/uuid"
)

// encodeTask 用于编码Task对象为json，用于存储到redis
func encodeTask(task config.Task) string {
	taskJSON, _ := json.Marshal(task)
	return string(taskJSON)
}

// decodeTask 用于将redis中的json编码转换为Task对象
func decodeTask(taskJSON string) config.Task {
	var task config.Task
	json.Unmarshal([]byte(taskJSON), &task)

	return task
}

// generateUUID 为任务生成一个唯一的ID
func generateUUID() string {
	return uuid.Generate().String()
}

// PushTask 用于将任务加入到Channel
func PushTask(task config.Task) (interface{}, error) {
	runtime := config.GetRuntime()
	if _, ok := runtime.Channels[task.Channel]; !ok {
		return nil, fmt.Errorf("task channel [%s] not exist", task.TaskName)
	}

	//  如果没有指定任务ID，则自动生成
	if task.ID == "" {
		task.ID = generateUUID()
	}

	client := createRedisClient()
	defer client.Close()

	log.Info("add task: %s -> %s", task.TaskName, task.Channel)

	return pushToQueueCmd.Run(
		client,
		[]string{TaskQueueKey(task.Channel), TaskQueueDistinctKey(task.Channel, task.TaskName)},
		encodeTask(task),
		runtime.Channels[task.Channel].Distinct,
	).Result()
}

// QueryTask function query task queue status
func QueryTask(channel string) (tasks []config.Task, err error) {
	client := createRedisClient()
	defer client.Close()

	tasks = []config.Task{}
	vals, err := client.LRange(TaskQueueKey(channel), 0, client.LLen(TaskQueueKey(channel)).Val()).Result()
	if err != nil {
		return
	}

	for _, v := range vals {
		task := decodeTask(v)

		status, _ := strconv.Atoi(client.Get(TaskQueueDistinctKey(task.Channel, task.TaskName)).Val())
		task.Status = "queued"
		if status != 1 {
			task.Status = "expired"
		}

		tasks = append(tasks, task)
	}

	// 查询执行中的任务
	for _, v := range client.HGetAll(TaskQueueExecKey(channel)).Val() {
		task := decodeTask(v)
		task.Status = "running"

		tasks = append(tasks, task)
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

		var task config.PrepareTask
		if err := json.Unmarshal([]byte(res[1]), &task); err == nil {
			PushTask(config.Task{
				TaskName: task.Name,
				Channel:  task.Channel,
			})
		}
	}
}
