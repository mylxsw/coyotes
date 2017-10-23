package log

import (
	"io"

	"github.com/mylxsw/coyotes/config"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("coyotes")

var DebugMode bool

// InitLogger 用来初始化日志输出文件
func InitLogger(logFile io.Writer, debugMode bool) {
	var format string
	if config.GetRuntime().Config.ColorfulTTY {
		format = "%{time:2006-01-02 15:04:05} %{color}[%{level}]%{color:reset} %{message}"
	} else {
		format = "%{time:2006-01-02 15:04:05} [%{level}] %{message}"
	}

	logging.SetBackend(
		logging.NewBackendFormatter(
			logging.NewLogBackend(logFile, "", 0),
			logging.MustStringFormatter(format),
		),
	)

	DebugMode = debugMode
}

func Info(message string, a ...interface{}) {
	log.Info(message, a...)
}

func Error(message string, a ...interface{}) {
	log.Error(message, a...)
}

func Debug(message string, a ...interface{}) {
	if !DebugMode {
		return
	}

	log.Debug(message, a...)
}

func Warning(message string, a ...interface{}) {
	log.Warning(message, a...)
}

func Notice(message string, a ...interface{}) {
	log.Notice(message, a...)
}
