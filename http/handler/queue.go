package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mylxsw/task-runner/http/response"
	"github.com/mylxsw/task-runner/scheduler"
)

// NewQueue function create new a task queue
func NewQueue(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

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
