package lanky_logger

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// config represents the configuration for the logger.
type config struct {
	isProduction     bool           // indicates whether the logger is running in production mode
	serviceName      string         // the name of the service using the logger
	additionalFields map[string]any // additional fields to include in the log messages
}

// Option is a function type that represents an option for configuring the logger.
// It takes a pointer to a config struct and modifies its properties.
type Option func(o *config)

// SetIsProduction sets the isProduction field of the config struct.
// It takes a boolean value as input and returns an Option function.
// The Option function modifies the config struct by setting the isProduction field to the given value.
func SetIsProduction(isProduction bool) Option {
	return func(o *config) {
		o.isProduction = isProduction
	}
}

// SetServiceName sets the service name for the logger configuration.
// It takes a string parameter `serviceName` and returns an `Option` function.
// The `Option` function modifies the `config` object by setting the `serviceName` field.
func SetServiceName(serviceName string) Option {
	return func(o *config) {
		o.serviceName = serviceName
	}
}

// SetFields sets additional fields for the logger configuration.
// It takes a map of string keys and any values as input.
// The additional fields will be included in the log output.
// Example usage:
//
//	logger.SetFields(map[string]interface{}{
//	  "key1": value1,
//	  "key2": value2,
//	})
func SetFields(fields map[string]any) Option {
	return func(o *config) {
		o.additionalFields = fields
	}
}

// NewInstance creates a new instance of the logrus.Logger with the provided options.
// It accepts a variadic parameter of Option functions that can be used to configure the logger.
// The default configuration includes:
// - isProduction: false
// - serviceName: "The Lanky Service"
// - additionalFields: an empty map[string]any
//
// Example usage:
//
//	log := NewInstance(WithProductionMode(true), WithServiceName("My Service"))
//
// Parameters:
// - opts: A variadic parameter of Option functions that can be used to configure the logger.
//
// Returns:
// - A pointer to the logrus.Logger instance.
func NewInstance(opts ...Option) *logrus.Logger {
	conf := &config{
		isProduction:     false,
		serviceName:      "The Lanky Service",
		additionalFields: make(map[string]any),
	}

	for _, opt := range opts {
		opt(conf)
	}

	var level logrus.Level

	if conf.isProduction {
		level = logrus.WarnLevel
	} else {
		level = logrus.DebugLevel
	}

	log := logrus.New()
	log.SetLevel(level)
	log.SetOutput(colorable.NewColorableStdout())
	log.AddHook(&defaultHookConfig{fields: conf.additionalFields})

	return log
}

type defaultHookConfig struct {
	fields map[string]any
}

func (dhc *defaultHookConfig) Fire(entry *logrus.Entry) error {
	for k, v := range dhc.fields {
		entry.Data[k] = v
	}
	return nil
}

func (dhc *defaultHookConfig) Levels() []logrus.Level {
	return logrus.AllLevels
}
