package channel

import (
	"fmt"
	"os"

	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/log"
)

// InitChannels init all channels
func InitChannels() {
	runtime := config.GetRuntime()

	// 初始化队列channel
	for key, ch := range GetChannels() {
		ch.Command = make(chan string, runtime.Config.ChannelCacheSize)
		runtime.Channels[key] = ch
	}
}

// GetChannels 获取broker中存储的所有channel
func GetChannels() map[string]*config.Channel {
	runtime := config.GetRuntime()

	channels, err := broker.GetTaskChannels()
	if err != nil {
		log.Error("get channels failed: %v", err)
		os.Exit(2)
	}

	// 默认三个channel： default, biz, cron
	for _, ch := range []string{runtime.Config.DefaultChannel, "biz", "cron"} {
		if _, ok := channels[ch]; ok {
			continue
		}

		channels[ch] = &config.Channel{
			Name:        ch,
			Distinct:    true,
			WorkerCount: runtime.Config.Concurrent,
		}

		broker.AddTaskChannel(channels[ch])
	}

	return channels
}

// NewChannel function create a new channel for task queue
func NewChannel(name string, distinct bool, workerCount int) (*config.Channel, error) {
	runtime := config.GetRuntime()

	if name == "" {
		return nil, fmt.Errorf("队列名称不能为空")
	}
	if _, ok := runtime.Channels[name]; ok {
		return nil, fmt.Errorf("任务队列 %s 已经存在", name)
	}

	channel := &config.Channel{
		Name:        name,
		Command:     make(chan string, runtime.Config.ChannelCacheSize),
		Distinct:    distinct,
		WorkerCount: workerCount,
	}

	if workerCount == 0 {
		channel.WorkerCount = runtime.Config.Concurrent
	}

	runtime.Channels[name] = channel
	// 将channel加入到broker存储，用于持久化
	broker.AddTaskChannel(channel)

	return channel, nil
}
