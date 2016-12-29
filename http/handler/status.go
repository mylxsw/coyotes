package handler

import (
	"fmt"
	"net/http"

	broker "github.com/mylxsw/task-runner/brokers/redis"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/http/response"
	"github.com/mylxsw/task-runner/log"
)

type taskResult struct {
	Tasks []broker.Task `json:"tasks"`
	Count int           `json:"count"`
}

// Status 查询所有channel的任务
func Status(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	runtime := config.GetRuntime()
	taskChannel := r.FormValue("channel")
	if taskChannel == "" {
		results := make(map[string]taskResult)

		for channelName := range runtime.Channels {
			tasks, err := broker.QueryTask(channelName)
			if err != nil {
				message := fmt.Sprintf("ERROR: %v", err)
				log.Error(message)
				w.Write(response.Failed(message))
				return
			}
			results[channelName] = taskResult{
				Tasks: tasks,
				Count: len(tasks),
			}
		}

		w.Write(response.Success(results))
	} else {
		tasks, err := broker.QueryTask(taskChannel)
		if err != nil {
			message := fmt.Sprintf("ERROR: %v", err)
			log.Error(message)
			w.Write(response.Failed(message))
			return
		}

		w.Write(response.Success(taskResult{
			Tasks: tasks,
			Count: len(tasks),
		}))
	}
}
