package gorequests

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, format string, v ...interface{})
}

type defaultLogger struct {
}

func (r *defaultLogger) Info(ctx context.Context, format string, v ...interface{}) {
	logrus.Infof(format, v...)
}

func (r *defaultLogger) Error(ctx context.Context, format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

var logger Logger = &defaultLogger{}

func SetLogger(l Logger) {
	logger = l
}
