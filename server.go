package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"

	"os"

	redisQueue "github.com/mylxsw/task-runner/queue/redis"
	redis "gopkg.in/redis.v5"
)

// Http Response
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

func startHTTPServer(runtime *config.Runtime) {
	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Config.Redis.Addr,
		Password: runtime.Config.Redis.Password,
		DB:       runtime.Config.Redis.DB,
	})
	defer client.Close()

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

		taskChannel := r.PostFormValue("channel")
		if taskChannel == "" {
			taskChannel = runtime.Config.DefaultChannel
		}

		tasks, err := redisQueue.QueryTaskQueue(client, taskChannel)
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Error(message)
			w.Write(failed(message))
			return
		}

		w.Write(success(struct {
			Tasks []redisQueue.Task
			Count int
		}{
			Tasks: tasks,
			Count: len(tasks),
		}))
	})

	// 任务推送
	http.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		taskName := r.PostFormValue("task")
		taskChannel := r.PostFormValue("channel")

		if taskChannel == "" {
			taskChannel = runtime.Config.DefaultChannel
		}

		if _, ok := runtime.Channels[taskChannel]; !ok {
			w.Write(failed("channel不存在"))
			return
		}

		rs, err := redisQueue.PushTaskToQueue(client, taskName, taskChannel)
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Error(message)
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

	log.Info("Http Listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.Http.ListenAddr))
	if err := http.ListenAndServe(runtime.Config.Http.ListenAddr, nil); err != nil {
		log.Error("Error: %v", err)
		os.Exit(2)
	}
}
