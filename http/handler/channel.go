package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/http/response"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/scheduler"
)

type taskResult struct {
	Tasks []brokers.Task `json:"tasks"`
	Count int            `json:"count"`
}

// StatusChannel 查询单个Channel的任务
func StatusChannel(w http.ResponseWriter, r *http.Request) {
	tasks, err := broker.GetTaskManager().QueryTask(mux.Vars(r)["channel_name"])
	if err != nil {
		message := fmt.Sprintf("error: %v", err)
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
	results := make(map[string]taskResult)
	for channelName := range config.GetRuntime().Channels {
		tasks, err := broker.GetTaskManager().QueryTask(channelName)
		if err != nil {
			message := fmt.Sprintf("error: %v", err)
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

// NewChannel function create new a task queue
func NewChannel(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	distinct := r.PostFormValue("distinct") == "true"
	workerCount, err := strconv.Atoi(r.PostFormValue("worker"))
	if err != nil {
		w.Write(response.Failed(fmt.Sprintf("字段workerCount不合法: %v", err)))
		return
	}

	err = scheduler.NewQueue(name, distinct, workerCount)
	if err != nil {
		w.Write(response.Failed(err.Error()))
		return
	}

	w.Write(response.Success(nil))
}

// RemoveChannel remove the spectified channel
func RemoveChannel(w http.ResponseWriter, r *http.Request) {

	err := broker.GetTaskManager().RemoveChannel(mux.Vars(r)["channel_name"])
	if err != nil {
		w.Write(response.Failed(fmt.Sprintf("删除失败：%v", err)))
		return
	}

	// TODO 需要检查channel中是否有task，是否有task在执行，以及关闭启动的worker

	w.Write(response.Success(nil))
}
