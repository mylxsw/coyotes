package handler

import (
	"fmt"
	"net/http"

	"strconv"

	"time"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/brokers"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/http/response"
	"github.com/mylxsw/coyotes/log"
)

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	// vars := mux.Vars(r)
	// channelName := vars["channel_name"]
	// taskName := vars["task_id"]

}

func PushTask(w http.ResponseWriter, r *http.Request) {
	response.SendJSONResponseHeader(w)

	taskName := r.PostFormValue("task")
	taskChannel := mux.Vars(r)["channel_name"]
	delaySec, _ := strconv.Atoi(r.PostFormValue("delay"))

	if taskChannel == "" {
		taskChannel = config.GetRuntime().Config.DefaultChannel
	}

	if _, ok := config.GetRuntime().Channels[taskChannel]; !ok {
		w.Write(response.Failed("channel不存在"))
		return
	}

	var taskID string
	var err error
	var existence bool

	if delaySec != 0 {
		taskID, existence, err = broker.GetTaskManager().AddDelayTask(
			time.Now().Add(time.Duration(delaySec)*time.Second),
			brokers.Task{
				TaskName: taskName,
				Channel:  taskChannel,
			})
	} else {
		taskID, existence, err = broker.GetTaskManager().AddTask(brokers.Task{
			TaskName: taskName,
			Channel:  taskChannel,
		})
	}

	if err != nil {
		message := fmt.Sprintf("failed push task [%s] to redis queue [%s]: %v", taskName, taskChannel, err)
		log.Error(message)
		w.Write(response.Failed(message))
		return
	}

	w.Write(response.Success(struct {
		TaskID   string `json:"task_id"`
		TaskName string `json:"task_name"`
		Result   bool   `json:"result"`
	}{
		TaskID:   taskID,
		TaskName: taskName,
		Result:   !existence,
	}))
}
