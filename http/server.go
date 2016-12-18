package http

import (
	"net/http"
	"os"

	"github.com/mylxsw/task-runner/config"
	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/http/handler"
	"github.com/mylxsw/task-runner/log"
)

// StartHTTPServer start an http server instance serving for api request
func StartHTTPServer() {

	// print welcome message
	http.HandleFunc("/", handler.Home)
	// check the server status
	http.HandleFunc("/status", handler.Status)
	// push task to task queue
	http.HandleFunc("/push", handler.TaskPush)

	runtime := config.GetRuntime()
	log.Debug("Http Listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))
	if err := http.ListenAndServe(runtime.Config.HTTP.ListenAddr, nil); err != nil {
		log.Error("Failed listening http on %s: %v", runtime.Config.HTTP.ListenAddr, err)
		os.Exit(2)
	}
}
