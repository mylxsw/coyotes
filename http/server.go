package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	redis "gopkg.in/redis.v5"
)

// Response is the result to user and it will be convert to a json object
type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// resposne function convert object to json response
func response(result Response) []byte {
	res, _ := json.Marshal(result)
	return res
}

// success function return a successful message to user
func success(result interface{}) []byte {
	return response(Response{
		StatusCode: 200,
		Message:    "ok",
		Data:       result,
	})
}

// failed function return a failed message to user
func failed(message string) []byte {
	return response(Response{
		StatusCode: 500,
		Message:    message,
	})
}

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer(runtime *config.Runtime) {
	client := redis.NewClient(&redis.Options{
		Addr:     runtime.Config.Redis.Addr,
		Password: runtime.Config.Redis.Password,
		DB:       runtime.Config.Redis.DB,
	})
	defer client.Close()

	// print welcome message
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(config.WelcomeMessage(runtime)))
	})

	// check the server status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		taskChannel := r.PostFormValue("channel")
		if taskChannel == "" {
			taskChannel = runtime.Config.DefaultChannel
		}

		tasks, err := broker.QueryTaskQueue(client, taskChannel)
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Error(message)
			w.Write(failed(message))
			return
		}

		w.Write(success(struct {
			Tasks []broker.Task `json:"tasks"`
			Count int           `json:"count"`
		}{
			Tasks: tasks,
			Count: len(tasks),
		}))
	})

	// push task to task queue
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

		rs, err := broker.PushTaskToQueue(client, taskName, taskChannel, runtime.Channels[taskChannel].Distinct)
		if err != nil {
			message := fmt.Sprintf("Failed push task [%s] to redis queue [%s]: %v", taskName, taskChannel, err)
			log.Error(message)
			w.Write(failed(message))
			return
		}

		w.Write(success(struct {
			TaskName string `json:"task_name"`
			Result   bool   `json:"result"`
		}{
			TaskName: taskName,
			Result:   int64(rs.(int64)) == 1,
		}))
	})

	log.Debug("Http Listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))
	if err := http.ListenAndServe(runtime.Config.HTTP.ListenAddr, nil); err != nil {
		log.Error("Failed listening http on %s: %v", runtime.Config.HTTP.ListenAddr, err)
		os.Exit(2)
	}
}
