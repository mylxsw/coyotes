package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strconv"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/task-runner/config"
	redis "gopkg.in/redis.v5"
)

type Task struct {
	TaskName string
	Status   int
}

type Response struct {
	StatusCode int
	Message    string
	Data       interface{}
}

func response(result Response) []byte {
	res, _ := json.Marshal(result)
	return res
}

func success(result interface{}) []byte {
	return response(Response{
		StatusCode: 200,
		Message:    "ok",
		Data:       result,
	})
}

func failed(message string) []byte {
	return response(Response{
		StatusCode: 500,
		Message:    message,
	})
}

func startHttpServer(runtime *config.Runtime) {
	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Redis.Addr,
		Password: runtime.Redis.Password,
		DB:       runtime.Redis.DB,
	})
	defer client.Close()

	script := `
local element = redis.call("EXISTS", KEYS[1])
if element ~= 1 then
    redis.call("LPUSH", "tasks:async:queue", ARGV[1])
	redis.call("SETEX", KEYS[1], 1800, '1')
end
return element
`
	pushToQueue := redis.NewScript(script)

	// 欢迎界面
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(welcomeMessage(runtime)))
	})

	// 运行状态查询
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		queueLen := client.LLen("tasks:async:queue").Val()
		vals, err := client.LRange("tasks:async:queue", 0, queueLen).Result()
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Println(message)
			w.Write(failed(message))
			return
		}

		tasks := []Task{}
		for _, v := range vals {
			status, _ := strconv.Atoi(client.Get(fmt.Sprintf("tasks:distinct:%s", v)).Val())
			tasks = append(tasks, Task{
				TaskName: v,
				// 0-去重key已过期，1-队列中，2-执行中
				Status: status,
			})
		}

		w.Write(success(struct {
			Tasks []Task
			Count int64
		}{
			Tasks: tasks,
			Count: queueLen,
		}))
	})

	// 任务推送
	http.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		taskName := r.PostFormValue("task")

		taskNameKey := fmt.Sprintf("tasks:distinct:%s", taskName)
		rs, err := pushToQueue.Run(client, []string{taskNameKey}, taskName).Result()
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Println(message)
			w.Write(failed(message))
			return
		}

		w.Write(success(struct {
			TaskName string
			Result   bool
		}{
			TaskName: taskName,
			Result:   int64(rs.(int64)) == 0,
		}))
	})

	log.Printf("Http Listening on %s", console.ColorfulText(console.TextCyan, runtime.Http.ListenAddr))
	if err := http.ListenAndServe(runtime.Http.ListenAddr, nil); err != nil {
		log.Fatalf("Error: %v", err)
	}
}