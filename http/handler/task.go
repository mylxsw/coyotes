package handler

import (
	"fmt"
	"net/http"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/http/response"
	"github.com/mylxsw/task-runner/log"
)

func TaskPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	runtime := config.GetRuntime()

	taskName := r.PostFormValue("task")
	taskChannel := r.PostFormValue("channel")

	if taskChannel == "" {
		taskChannel = runtime.Config.DefaultChannel
	}

	if _, ok := runtime.Channels[taskChannel]; !ok {
		w.Write(response.Failed("channel不存在"))
		return
	}

	rs, err := broker.PushTask(taskName, taskChannel, runtime.Channels[taskChannel].Distinct)
	if err != nil {
		message := fmt.Sprintf("Failed push task [%s] to redis queue [%s]: %v", taskName, taskChannel, err)
		log.Error(message)
		w.Write(response.Failed(message))
		return
	}

	w.Write(response.Success(struct {
		TaskName string `json:"task_name"`
		Result   bool   `json:"result"`
	}{
		TaskName: taskName,
		Result:   int64(rs.(int64)) == 1,
	}))
}
