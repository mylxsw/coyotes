package redis

import (
	"github.com/mylxsw/task-runner/config"
	"gopkg.in/redis.v5"
)

var runtime *config.Runtime

func init() {
	runtime = config.GetRuntime()
}

type Task struct {
	TaskName string `json:"task_name"`
	Status   int    `json:"status"`
}

var pushToQueueCmd = redis.NewScript(`
-- KEYS[1]=队列key
-- KEYS[2]=去重key
-- ARGV[1]=命令
-- ARGV[2]=是否启用去重

-- redis.log(redis.LOG_NOTICE, type(ARGV[2]) .. " : " .. ARGV[2])

-- 如果不启用去重复功能，则直接push到任务队列
if ARGV[2] == '0' then 
	redis.call("LPUSH", KEYS[1], ARGV[1])
	return 1
end

-- 不启用去重复功能，先判断是否存在去重key，存在则不添加队列
-- 不存在则添加队列并设置去重key，有效期1800s
local element = redis.call("EXISTS", KEYS[2])
if element ~= 1 then
	redis.call("LPUSH", KEYS[1], ARGV[1])
	redis.call("SETEX", KEYS[2], 1800, '1')
	
	return 1
end

return 0
`)

func createRedisClient() *redis.Client {
	runtime := config.GetRuntime()
	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Config.Redis.Addr,
		Password: runtime.Config.Redis.Password,
		DB:       runtime.Config.Redis.DB,
	})

	return client
}
