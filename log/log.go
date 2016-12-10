package log

import (
	"fmt"
	syslog "log"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelError = "ERROR"
)

func Info(message string, a ...interface{}) {
	writeLog(LevelInfo, message, a...)
}

func Error(message string, a ...interface{}) {
	writeLog(LevelError, message, a...)
}

func Debug(message string, a ...interface{}) {
	writeLog(LevelDebug, message, a...)
}

func writeLog(level string, message string, a ...interface{}) {
	syslog.Printf("[%s] %s", level, fmt.Sprintf(message, a...))
}
