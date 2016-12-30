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
	r.HandleFunc("/status", handler.Status).Methods("GET")
	r.HandleFunc("/push", handler.PushTask).Methods("POST")
	r.HandleFunc("/queue", handler.NewQueue).Methods("POST")

	log.Debug("http listening on %s", console.ColorfulText(console.TextCyan, runtime.Config.HTTP.ListenAddr))

	if err := http.ListenAndServe(runtime.Config.HTTP.ListenAddr, r); err != nil {
		log.Error("failed listening http on %s: %v", runtime.Config.HTTP.ListenAddr, err)
		os.Exit(2)
	}
}
