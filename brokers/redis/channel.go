package redis

import (
	"encoding/json"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
)

// GetTaskChannel get a channel from redis
func GetTaskChannel(channelName string) (channel config.Channel, err error) {
	client := createRedisClient()
	defer client.Close()

	result, err := client.HGet(TaskChannelsKey(), channelName).Result()
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

// GetTaskChannels return all channels
func GetTaskChannels() (channels map[string]*config.Channel, err error) {
	channels = make(map[string]*config.Channel)

	client := createRedisClient()
	defer client.Close()

	results, err := client.HGetAll(TaskChannelsKey()).Result()
	if err != nil {
		return nil, err
	}

	for key, res := range results {
		channel := config.Channel{}
		err = json.Unmarshal([]byte(res), &channel)
		if err != nil {
			log.Error("parse channel [%s] from json to object failed", key)
			continue
		}

		channels[key] = &channel
	}

	return
}

// AddTaskChannel add a channel to redis for persistence
func AddTaskChannel(channel *config.Channel) error {
	client := createRedisClient()
	defer client.Close()

	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return err
	}

	err = client.HSet(TaskChannelsKey(), channel.Name, string(channelJSON)).Err()
	if err != nil {
		return err
	}

	return nil
}

// RemoveTaskChannel remove a channel from redis
func RemoveTaskChannel(channelName string) error {
	client := createRedisClient()
	defer client.Close()

	err := client.HDel(TaskChannelsKey(), channelName).Err()
	if err != nil {
		return err
	}

	return nil
}
