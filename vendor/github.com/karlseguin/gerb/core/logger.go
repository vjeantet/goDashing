package core

import (
	"log"
	"os"
)

type Logger interface {
	Error(v ...interface{})
}

var Log Logger = &DefaultLogger{log.New(os.Stdout, "[gerb] ", log.Ldate|log.Ltime)}

type DefaultLogger struct {
	l *log.Logger
}

func (l *DefaultLogger) Error(v ...interface{}) {
	l.l.Println(v...)
}

type SilentLogger struct{}

func (l *SilentLogger) Error(v ...interface{}) {}
