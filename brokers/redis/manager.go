package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"context"

	"github.com/docker/distribution/uuid"
	"github.com/mylxsw/coyotes/brokers"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
	redis "gopkg.in/redis.v5"
)

// TaskManager 任务管理器
type TaskManager struct {
	runtime *config.Runtime
	client  *redis.Client
}

var brokerManager *TaskManager

// GetTaskManager Create a redis Manager
func GetTaskManager() *TaskManager {
	if brokerManager == nil {
		brokerManager = &TaskManager{
			client:  createRedisClient(),
			runtime: config.GetRuntime(),
		}
	}

	return brokerManager
}

// Close manager
func (manager *TaskManager) Close() {
	manager.client.Close()
}

// AddTask 用于将任务加入到Channel
func (manager *TaskManager) AddTask(task brokers.Task) (id string, existence bool, err error) {
	if _, ok := manager.runtime.Channels[task.Channel]; !ok {
		return "", false, fmt.Errorf("task channel [%s] not exist", task.TaskName)
	}

	//  如果没有指定任务ID，则自动生成
	if task.ID == "" {
		task.ID = generateUUID()
	}

	log.Info("add task: %s -> %s", task.TaskName, task.Channel)

	val, err := pushToQueueCmd.Run(
		manager.client,
		[]string{TaskQueueKey(task.Channel), TaskQueueDistinctKey(task.Channel, task.TaskName)},
		encodeTask(task),
		manager.runtime.Channels[task.Channel].Distinct,
	).Result()

	if err != nil {
		return "", false, err
	}

	return task.ID, int64(val.(int64)) != 1, nil
}

// QueryTask function query task queue status
func (manager *TaskManager) QueryTask(channel string) (tasks []brokers.Task, err error) {

	tasks = []brokers.Task{}
	vals, err := manager.client.LRange(TaskQueueKey(channel), 0, manager.client.LLen(TaskQueueKey(channel)).Val()).Result()
	if err != nil {
		return
	}

	for _, v := range vals {
		task := decodeTask(v)

		status, _ := strconv.Atoi(manager.client.Get(TaskQueueDistinctKey(task.Channel, task.TaskName)).Val())
		task.Status = "queued"
		if status != 1 {
			task.Status = "expired"
		}

		tasks = append(tasks, task)
	}

	// 查询执行中的任务
	for _, v := range manager.client.HGetAll(TaskQueueExecKey(channel)).Val() {
		task := decodeTask(v)
		task.Status = "running"

		tasks = append(tasks, task)
	}

	return
}

// GetTaskChannel 从Redis中查询某个channel信息
func (manager *TaskManager) GetTaskChannel(channelName string) (channel brokers.Channel, err error) {

	result, err := manager.client.HGet(TaskChannelsKey(), channelName).Result()
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(result), &channel)
	if err != nil {
		log.Error("parse channel [%s] from json to object failed", channelName)
		return
	}

	return
}

// GetTaskChannels 返回偶有的channel信息
func (manager *TaskManager) GetTaskChannels() (channels map[string]*brokers.Channel, err error) {
	channels = make(map[string]*brokers.Channel)

	results, err := manager.client.HGetAll(TaskChannelsKey()).Result()
	if err != nil {
		return nil, err
	}

	for key, res := range results {
		channel := brokers.Channel{}
		err = json.Unmarshal([]byte(res), &channel)
		if err != nil {
			log.Error("parse channel [%s] from json to object failed", key)
			continue
		}

		channels[key] = &channel
	}

	return
}

// AddChannel 新增一个channel，会持久化到Redis中
func (manager *TaskManager) AddChannel(channel *brokers.Channel) error {
	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return err
	}

	err = manager.client.HSet(TaskChannelsKey(), channel.Name, string(channelJSON)).Err()
	if err != nil {
		return err
	}

	return nil
}

// RemoveChannel 从Redis中移除channel
func (manager *TaskManager) RemoveChannel(channelName string) error {
	err := manager.client.HDel(TaskChannelsKey(), channelName).Err()
	if err != nil {
		return err
	}

	return nil
}

// encodeTask 用于编码Task对象为json，用于存储到redis
func encodeTask(task brokers.Task) string {
	taskJSON, _ := json.Marshal(task)
	return string(taskJSON)
}

// decodeTask 用于将redis中的json编码转换为Task对象
func decodeTask(taskJSON string) brokers.Task {
	var task brokers.Task
	json.Unmarshal([]byte(taskJSON), &task)

	return task
}

// generateUUID 为任务生成一个唯一的ID
func generateUUID() string {
	return uuid.Generate().String()
}

// TransferPrepareTask 将prepare队列中的任务加入正式的任务队列
func TransferPrepareTask(ctx context.Context) {
	client := createRedisClient()
	defer client.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := client.BRPop(2*time.Second, TaskPrepareQueueKey()).Result()
			if err != nil {
				continue
			}

			var task brokers.PrepareTask
			if err := json.Unmarshal([]byte(res[1]), &task); err == nil {
				GetTaskManager().AddTask(brokers.Task{
					TaskName: task.Name,
					Channel:  task.Channel,
					Command:  task.Command,
				})
			}
		}
	}
}
