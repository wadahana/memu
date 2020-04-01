package log

import (
	"io"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("memu")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func initLogger(output io.Writer) {

	backend := logging.NewLogBackend(output, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
    backendLeveled.SetLevel(logging.DEBUG, "")
    logging.SetBackend(backendLeveled)
}	

func Print(i ...interface{}) {
	logger.Print(i...)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}


func Debug(i ...interface{}) {
	logger.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(i ...interface{}) {
	logger.Info(i...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(i ...interface{}) {
	logger.Warn(i...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(i ...interface{}) {
	logger.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	logger.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panic(i ...interface{}) {
	logger.Panic(i...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}