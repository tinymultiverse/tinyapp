/*
Copyright 2024 BlackRock, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerType is a constant string to refer to the type of logger to create.
type LoggerType string

const (
	// Development refers to the development logger.
	Development LoggerType = "development"

	// Production refers to the production logger.
	Production LoggerType = "production"
)

func getLoggerTypeFromEnvironment() LoggerType {
	logType := os.Getenv("LOG_TYPE")
	switch strings.ToLower(logType) {
	case string(Production):
		return Production
	case string(Development):
		return Development
	default:
		return Development
	}
}

type LogOption func(config zap.Config) zap.Config

func WithNoCaller() LogOption {
	return func(config zap.Config) zap.Config {
		config.DisableCaller = true
		return config
	}
}

func WithNoColor() LogOption {
	return func(config zap.Config) zap.Config {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return config
	}
}

func WithOutputPaths(outputPaths []string) LogOption {
	return func(config zap.Config) zap.Config {
		config.OutputPaths = outputPaths
		return config
	}
}

// InitLoggerFromEnvironment looks at the environment to determine log level.
func InitLoggerFromEnvironment(options ...LogOption) {
	loggerTypeFromEnvironment := getLoggerTypeFromEnvironment()
	InitLoggerExplicit(loggerTypeFromEnvironment, options...)
}

// InitLoggerExplicit will initialize a logger based on the explicit type passed in from the caller.
func InitLoggerExplicit(loggerType LoggerType, options ...LogOption) {
	var mainLogger *zap.Logger
	var err error
	switch loggerType {
	case Development:
		mainLogger, err = getDevelopmentLogger(options...)
	case Production:
		mainLogger, err = getProductionLogger(options...)
	}
	if err != nil {
		log.Panic("could not initialize logger", err)
	}
	zap.ReplaceGlobals(mainLogger)
}

// RegisterHook registers a hook function and replaces the global logger with a new logger using the updated zap core.
func RegisterHook(hook func(zapcore.Entry) error) {
	core := zap.L().Core()
	core = zapcore.RegisterHooks(core, hook)
	newLogger := zap.New(core)
	zap.ReplaceGlobals(newLogger)
}

func getDevelopmentLogger(options ...LogOption) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	config.OutputPaths = []string{"stderr"}

	for _, option := range options {
		config = option(config)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func getProductionLogger(options ...LogOption) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.OutputPaths = []string{"stderr"}

	for _, option := range options {
		config = option(config)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
