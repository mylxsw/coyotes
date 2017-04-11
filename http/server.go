package http

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/http/handler"
	"github.com/mylxsw/coyotes/log"
)

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer() {
	runtime := config.GetRuntime()

	r := mux.NewRouter()

	r.HandleFunc("/", handler.Home).Methods("GET")

	// 查看所有channel的状态
	r.HandleFunc("/channels", handler.StatusChannels).Methods("GET")
	// 创建新的channel
	r.HandleFunc("/channels", handler.NewChannel).Methods("POST")
	// 查看某个channel的状态
	r.HandleFunc("/channels/{channel_name}", handler.StatusChannel).Methods("GET")
	// 删除某个channel
	r.HandleFunc("/channels/{channel_name}", handler.RemoveChannel).Methods("DELETE")

	// 推送新的task到channel
	r.HandleFunc("/channels/{channel_name}/tasks", handler.PushTask).Methods("POST")
	r.HandleFunc("/channels/{channel_name}/tasks/{task_id}", handler.RemoveTask).Methods("DELETE")

	log.Debug("http listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))

	if err := http.ListenAndServe(runtime.Config.HTTP.ListenAddr, r); err != nil {
		log.Error("failed listening http on %s: %v", runtime.Config.HTTP.ListenAddr, err)
		os.Exit(2)
	}
}
