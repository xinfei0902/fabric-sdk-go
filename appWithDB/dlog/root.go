package dlog

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
}

// Global Default Values
var (
	GlobalConsoleMark  = "console"
	GlobalDebugLevel   = logrus.DebugLevel
	GlobalDefaultLevel = logrus.InfoLevel
)

var globalFileHolder *os.File

// InitLog to start log
func InitLog(name string, level string, json bool) (err error) {
	if name == GlobalConsoleMark {
		InitConsole(logrus.DebugLevel, json)
		return
	}

	err = CloseLog()
	if err != nil {
		return
	}

	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}

	globalFileHolder = f

	var base = logrus.InfoLevel
	switch level {
	case logrus.DebugLevel.String():
		base = logrus.DebugLevel
	case logrus.InfoLevel.String():
		base = logrus.InfoLevel
	case logrus.WarnLevel.String():
		base = logrus.WarnLevel
	case logrus.ErrorLevel.String():
		base = logrus.ErrorLevel
	case logrus.FatalLevel.String():
		base = logrus.FatalLevel
	default:
		base = GlobalDefaultLevel
	}

	InitIO(globalFileHolder, base, json)

	return
}

// CloseLog to end log
func CloseLog() (err error) {
	if globalFileHolder != nil {
		err = globalFileHolder.Close()
		globalFileHolder = nil
	}
	return
}

// InitConsole log output into console
func InitConsole(level logrus.Level, json bool) {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)
	setFormat(level, json)
}

// InitIO log output into IO
func InitIO(io io.Writer, level logrus.Level, json bool) {
	logrus.SetLevel(level)
	logrus.SetOutput(io)
	setFormat(level, json)
}

func setFormat(level logrus.Level, json bool) {
	logrus.SetLevel(level)

	if json {
		// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}

func DebugLog(name string, serial uint64) *logrus.Entry {
	return logrus.WithField("model", name).WithField("serial", serial)
}

func Debug(msg string) {
	logrus.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debug(fmt.Sprintf(format, args...))
}

func Info(msg string) {
	logrus.Info(msg)
}

func Warn(msg string) {
	logrus.Warn(msg)
}
func Warnf(format string, args ...interface{}) {
	logrus.Warn(fmt.Sprintf(format, args...))
}

func Error(err error) {
	logrus.Error(err)
}

func Fatal(msg string) {
	logrus.Fatal(msg)
}
