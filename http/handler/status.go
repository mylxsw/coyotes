package handler

import (
	"fmt"
	"net/http"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	"github.com/mylxsw/task-runner/http/response"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/log"
)

func Status(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	runtime := config.GetRuntime()

	taskChannel := r.PostFormValue("channel")
	if taskChannel == "" {
		taskChannel = runtime.Config.DefaultChannel
	}

	tasks, err := broker.QueryTask(taskChannel)
	if err != nil {
		message := fmt.Sprintf("ERROR: %v", err)
		log.Error(message)
		w.Write(response.Failed(message))
		return
	}

	w.Write(response.Success(struct {
		Tasks []broker.Task `json:"tasks"`
		Count int           `json:"count"`
	}{
		Tasks: tasks,
		Count: len(tasks),
	}))
}
