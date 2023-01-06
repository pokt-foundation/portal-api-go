package logger

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrEmptyService when creating a new instance the service value is an empty string
	ErrEmptyService = errors.New("the service name is empty")

	log = logrus.New()
)

func init() {
	log.Formatter = &logrus.JSONFormatter{}
}

// Logger data handler
type Logger struct {
	ServiceName string
	Output      io.Writer
	LogObject
}

type LoggerInterface interface {
	LogName() string
	LogProperties() map[string]interface{}
}

// New return a new Logger instance
func New(service string) *Logger {
	if service == "" {
		panic(ErrEmptyService)
	}

	log.Formatter = &logrus.JSONFormatter{}

	output := io.Discard

	log.SetOutput(output)

	return &Logger{
		ServiceName: service,
		Output:      output,
	}
}

func mapObjectsToProperties(objects []LoggerInterface) map[string]interface{} {
	properties := map[string]interface{}{}

	for _, object := range objects {
		properties[object.LogName()] = object.LogProperties()
	}

	return properties
}

// Log generate the log with the given parameters y return in stderr a json.
func (logger *Logger) Log(ctx context.Context, eventName, message string, level logrus.Level, properties map[string]interface{}) {
	now := time.Now()

	output := map[string]interface{}{
		"event":      eventName,
		"level":      level,
		"service":    logger.ServiceName,
		"properties": properties,
		"message":    message,
		"time":       now.Format(time.RFC3339Nano),
	}

	log.WithFields(output).Log(level)

}

func (logger *Logger) Info(ctx context.Context, eventName, message string, objects []LoggerInterface) {
	logger.Log(ctx, eventName, message, logrus.InfoLevel, mapObjectsToProperties(objects))
}

func (logger *Logger) Error(ctx context.Context, eventName, message string, objects []LoggerInterface) {
	logger.Log(ctx, eventName, message, logrus.ErrorLevel, mapObjectsToProperties(objects))
}
