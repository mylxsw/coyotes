package handler

import (
	"net/http"

	"github.com/mylxsw/coyotes/config"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(config.WelcomeMessage()))
}
