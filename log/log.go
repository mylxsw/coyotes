package log

import (
	"fmt"
	syslog "log"

	"github.com/mylxsw/task-runner/console"
)

const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelError   = "ERROR"
	LevelWarning = "WARNING"
)

func init() {
}

func Info(message string, a ...interface{}) {
	writeLog(LevelInfo, message, a...)
}

func Error(message string, a ...interface{}) {
	writeLog(LevelError, message, a...)
}

func Debug(message string, a ...interface{}) {
	writeLog(LevelDebug, message, a...)
}

func Warning(message string, a ...interface{}) {
	writeLog(LevelWarning, message, a...)
}

func writeLog(level string, message string, a ...interface{}) {
	syslog.Printf("[%s] %s", console.ColorfulText(console.TextBlue, level), fmt.Sprintf(message, a...))
}
