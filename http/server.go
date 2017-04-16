package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/http/handler"
	"github.com/mylxsw/coyotes/log"
	"context"
)

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer(ctx context.Context) {
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

	srv := &http.Server{
		Addr: runtime.Config.HTTP.ListenAddr,
		Handler: r,
	}

	go func() {
		select {
		case <- ctx.Done():
			srv.Shutdown(ctx)
		}
	}()

	log.Debug("http listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))
	if err := srv.ListenAndServe(); err != nil {
		log.Warning("http server stopped: %v", err)
	}
}
