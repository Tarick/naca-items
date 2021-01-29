package tracing

import (
	"fmt"

	"go.uber.org/zap"
)

// zapLogger is zap logger implementation of jaeger.Logger
// logger delegates all calls to the underlying zap.Logger
type zapLogger struct {
	logger *zap.SugaredLogger
}

// Info logs an info msg
func (l zapLogger) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

// Error logs an error msg
func (l zapLogger) Error(msg string) {
	l.logger.Error(msg)
}

// Debugf logs an debug msg
func (l zapLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func NewZapLogger(logger *zap.SugaredLogger) *zapLogger {
	return &zapLogger{logger: logger}
}
