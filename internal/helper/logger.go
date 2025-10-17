package helper

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	*log.Logger
}

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s]", timestamp, level)

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	l.Logger.Printf("%s %s", prefix, message)
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}
