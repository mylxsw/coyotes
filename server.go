package main

import (
	"log"
	"net/http"

	"encoding/json"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/task-runner/config"
)

func startHttpServer(runtime *config.Runtime) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(welcomeMessage()))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {

	})

	http.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		task := r.PostFormValue("task")

		res, _ := json.Marshal(struct {
			TaskName string
		}{
			TaskName: task,
		})

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write(res)
	})

	log.Printf("Http Listening on %s", console.ColorfulText(console.TextCyan, runtime.Http.ListenAddr))
	if err := http.ListenAndServe(runtime.Http.ListenAddr, nil); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
