package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/http/response"
)

// GetDelayTasks 查询所有延迟任务
func GetDelayTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := broker.GetTaskManager().GetDelayTasks()
	if err != nil {
		w.Write(response.Failed("查询失败"))
		return
	}

	w.Write(response.Success(tasks))
}

// GetDelayTask Get specified delay task
func GetDelayTask(w http.ResponseWriter, r *http.Request) {
	task, err := broker.GetTaskManager().GetDelayTask(mux.Vars(r)["task_id"])
	if err != nil {
		w.Write(response.Failed(err.Error()))
		return
	}

	w.Write(response.Success(task))
}

// RemoveDelayTask Remove specified delay task from delay queue
func RemoveDelayTask(w http.ResponseWriter, r *http.Request) {
	task, err := broker.GetTaskManager().RemoveDelayTask(mux.Vars(r)["task_id"])
	if err != nil {
		w.Write(response.Failed(err.Error()))
		return
	}

	w.Write(response.Success(task))
}
