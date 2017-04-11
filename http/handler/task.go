package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	broker "github.com/mylxsw/coyotes/brokers/redis"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/http/response"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/brokers"
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

	if taskChannel == "" {
		taskChannel = config.GetRuntime().Config.DefaultChannel
	}

	if _, ok := config.GetRuntime().Channels[taskChannel]; !ok {
		w.Write(response.Failed("channel不存在"))
		return
	}

	rs, err := broker.PushTask(brokers.Task{
		TaskName: taskName,
		Channel:  taskChannel,
	})
	if err != nil {
		message := fmt.Sprintf("failed push task [%s] to redis queue [%s]: %v", taskName, taskChannel, err)
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
