package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger levels: info, debug, error
const (
	InfoLevel  = "info"
	DebugLevel = "debug"
	ErrorLevel = "error"
)

var globalLogger *zap.Logger

// Init initializes the global logger with the specified level
func Init(level string) error {
	config := zap.NewDevelopmentConfig()

	// Set log level based on input
	switch level {
	case DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configure output format for CLI
	config.Development = false
	config.Encoding = "console"
	config.OutputPaths = []string{"stderr"}      // Log to stderr instead of stdout
	config.ErrorOutputPaths = []string{"stderr"} // Error logs to stderr
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		FunctionKey:    "",
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to no-op logger if not initialized
		globalLogger = zap.NewNop()
	}
	return globalLogger
}

// Info logs an info level message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Debug logs a debug level message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Error logs an error level message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}
