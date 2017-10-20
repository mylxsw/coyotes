package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/http/response"
)

// FailedTasksInChannel 查询channel的失败队列中所有的任务
func FailedTasksInChannel(w http.ResponseWriter, r *http.Request) {
	channelName := mux.Vars(r)["channel_name"]
	tasks, err := broker.GetTaskManager().GetFailedTasks(channelName)
	if err != nil {
		w.Write(response.Failed(fmt.Sprintf("查询失败:%v", err)))
		return
	}

	w.Write(response.Success(tasks))
}

// GetFailedTask 查询某个失败的任务
func GetFailedTask(w http.ResponseWriter, r *http.Request) {
	taskChannel := mux.Vars(r)["channel_name"]
	taskID := mux.Vars(r)["task_id"]

	var task brokers.Task
	var err error
	if task, err = broker.GetTaskManager().GetFailedTask(taskChannel, taskID); err != nil {
		w.Write(response.Failed("任务不存在"))
		return
	}

	w.Write(response.Success(task))
}

// RetryTask 重试失败的任务
func RetryTask(w http.ResponseWriter, r *http.Request) {
	taskChannel := mux.Vars(r)["channel_name"]
	taskID := mux.Vars(r)["task_id"]

	if err := broker.GetTaskManager().RetryFailedTask(taskChannel, taskID); err != nil {
		w.Write(response.Failed("任务不存在"))
		return
	}

	w.Write(response.Success(nil))
}

// RemoveFailedTask Remove failed task
func RemoveFailedTask(w http.ResponseWriter, r *http.Request) {
	taskChannel := mux.Vars(r)["channel_name"]
	taskID := mux.Vars(r)["task_id"]

	var task brokers.Task
	var err error

	if task, err = broker.GetTaskManager().RemoveFailedTask(taskChannel, taskID); err != nil {
		w.Write(response.Failed("任务不存在"))
		return
	}

	w.Write(response.Success(task))
}
