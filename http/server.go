package http

import (
	"net/http"

	"context"

	"github.com/gorilla/mux"
	"github.com/mylxsw/coyotes/config"
	"github.com/mylxsw/coyotes/http/handler"
	mw "github.com/mylxsw/coyotes/http/middleware"
	"github.com/mylxsw/coyotes/log"
)

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer(ctx context.Context) {
	runtime := config.GetRuntime()

	r := mux.NewRouter()

	r.HandleFunc("/", mw.Handler(handler.Home, mw.WithHTMLResponse)).Methods("GET")

	// 查看所有channel的状态
	r.HandleFunc("/channels", mw.Handler(handler.StatusChannels, mw.WithJSONResponse)).Methods("GET")
	// 创建新的channel
	r.HandleFunc("/channels", mw.Handler(handler.NewChannel, mw.WithJSONResponse)).Methods("POST")
	// 查看某个channel的状态
	r.HandleFunc("/channels/{channel_name}", mw.Handler(handler.StatusChannel, mw.WithJSONResponse)).Methods("GET")
	// 删除某个channel
	r.HandleFunc("/channels/{channel_name}", mw.Handler(handler.RemoveChannel, mw.WithJSONResponse)).Methods("DELETE")

	// 推送新的task到channel
	r.HandleFunc("/channels/{channel_name}/tasks", mw.Handler(handler.PushTask, mw.WithJSONResponse)).Methods("POST")
	r.HandleFunc("/channels/{channel_name}/tasks/{task_id}", mw.Handler(handler.RemoveTask, mw.WithJSONResponse)).Methods("DELETE")

	// 重试失败的任务
	r.HandleFunc("/channels/{channel_name}/failed-tasks", mw.Handler(handler.FailedTasksInChannel, mw.WithJSONResponse)).Methods("GET")
	r.HandleFunc("/channels/{channel_name}/failed-tasks/{task_id}", mw.Handler(handler.GetFailedTask, mw.WithJSONResponse)).Methods("GET")
	r.HandleFunc("/channels/{channel_name}/failed-tasks/{task_id}", mw.Handler(handler.RetryTask, mw.WithJSONResponse)).Methods("POST")

	srv := &http.Server{
		Addr:    runtime.Config.HTTP.ListenAddr,
		Handler: r,
	}

	go func() {
		select {
		case <-ctx.Done():
			srv.Shutdown(ctx)
		}
	}()

	log.Debug("http listening on %s", runtime.Config.HTTP.ListenAddr)
	if err := srv.ListenAndServe(); err != nil {
		log.Warning("http server stopped: %v", err)
	}
}
