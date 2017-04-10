package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	broker "github.com/mylxsw/task-runner/brokers/redis"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/http/response"
	"github.com/mylxsw/task-runner/log"
)

type taskResult struct {
	Tasks []broker.Task `json:"tasks"`
	Count int           `json:"count"`
}

// StatusChannel 查询单个Channel的任务
func StatusChannel(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	tasks, err := broker.QueryTask(mux.Vars(r)["channel_name"])
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

// StatusChannels 查询所有channel的任务
func StatusChannels(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	results := make(map[string]taskResult)
	for channelName := range config.GetRuntime().Channels {
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
}
