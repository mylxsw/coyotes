package http

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/http/handler"
	"github.com/mylxsw/task-runner/log"
)

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer() {
	runtime := config.GetRuntime()

	r := mux.NewRouter()

	r.HandleFunc("/", handler.Home).Methods("GET")
	// 查看所有channel的状态
	r.HandleFunc("/channels", handler.StatusChannels).Methods("GET")
	// 创建新的channel
	r.HandleFunc("/channels", handler.NewQueue).Methods("POST")
	// 查看某个channel的状态
	r.HandleFunc("/channels/{channel_name}", handler.StatusChannel).Methods("GET")
	// 推送新的task到channel
	r.HandleFunc("/channels/{channel_name}", handler.PushTask).Methods("POST")

	log.Debug("http listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))

	if err := http.ListenAndServe(runtime.Config.HTTP.ListenAddr, r); err != nil {
		log.Error("failed listening http on %s: %v", runtime.Config.HTTP.ListenAddr, err)
		os.Exit(2)
	}
}
