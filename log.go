package gorequests

import (
	"context"
	"log"
	"os"
)

type Logger interface {
	Info(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, format string, v ...interface{})
}

type defaultLogger struct {
	logger *log.Logger
}

func (r *defaultLogger) Info(ctx context.Context, format string, v ...interface{}) {
	r.logger.Printf(format, v...)
}

func (r *defaultLogger) Error(ctx context.Context, format string, v ...interface{}) {
	r.logger.Printf(format, v...)
}

func newDefaultLogger() Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "[gorequests] ", log.LstdFlags),
	}
}
