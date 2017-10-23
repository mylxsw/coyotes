package scheduler

import (
	"fmt"
	"os"

	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
)

// InitChannels init all channels
func InitChannels() {
	runtime := config.GetRuntime()

	// 初始化队列channel
	for key, ch := range GetChannels() {
		ch.Task = make(chan brokers.Task, runtime.Config.ChannelCacheSize)
		runtime.Channels[key] = ch
	}
}

// GetChannels 获取broker中存储的所有channel
func GetChannels() map[string]*brokers.Channel {
	runtime := config.GetRuntime()

	channels, err := broker.GetTaskManager().GetTaskChannels()
	if err != nil {
		log.Error("get channels failed: %v", err)
		os.Exit(2)
	}

	// 默认三个channel： default, biz, cron, delayed
	for _, ch := range []string{runtime.Config.DefaultChannel, "biz", "cron", "delayed"} {
		if _, ok := channels[ch]; ok {
			continue
		}

		channels[ch] = &brokers.Channel{
			Name:        ch,
			Distinct:    true,
			WorkerCount: runtime.Config.Concurrent,
		}

		broker.GetTaskManager().AddChannel(channels[ch])
	}

	return channels
}

// NewChannel function create a new channel for task queue
func NewChannel(name string, distinct bool, workerCount int) (*brokers.Channel, error) {
	runtime := config.GetRuntime()

	if name == "" {
		return nil, fmt.Errorf("channel name required")
	}
	if _, ok := runtime.Channels[name]; ok {
		return nil, fmt.Errorf("channel %s has existed", name)
	}

	channel := &brokers.Channel{
		Name:        name,
		Task:        make(chan brokers.Task, runtime.Config.ChannelCacheSize),
		Distinct:    distinct,
		WorkerCount: workerCount,
	}

	if workerCount == 0 {
		channel.WorkerCount = runtime.Config.Concurrent
	}

	runtime.Channels[name] = channel
	// 将channel加入到broker存储，用于持久化
	broker.GetTaskManager().AddChannel(channel)

	return channel, nil
}
