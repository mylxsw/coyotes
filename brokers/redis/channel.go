package redis

import (
	"encoding/json"

	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/brokers"
)

// GetTaskChannel 从Redis中查询某个channel信息
func GetTaskChannel(channelName string) (channel brokers.Channel, err error) {
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

// GetTaskChannels 返回偶有的channel信息
func GetTaskChannels() (channels map[string]*brokers.Channel, err error) {
	channels = make(map[string]*brokers.Channel)

	client := createRedisClient()
	defer client.Close()

	results, err := client.HGetAll(TaskChannelsKey()).Result()
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
func AddChannel(channel *brokers.Channel) error {
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

// RemoveChannel 从Redis中移除channel
func RemoveChannel(channelName string) error {
	client := createRedisClient()
	defer client.Close()

	err := client.HDel(TaskChannelsKey(), channelName).Err()
	if err != nil {
		return err
	}

	return nil
}
