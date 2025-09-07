package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	InfoLevel  = "info"
	DebugLevel = "debug"
	ErrorLevel = "error"
)

var globalLogger *zap.Logger

func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}

func Init(level string) error {
	config := zap.NewDevelopmentConfig()

	switch level {
	case DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	}

	config.Development = false
	config.Encoding = "console"
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

func Get() *zap.Logger {
	if globalLogger == nil {
		globalLogger = zap.NewNop()
	}
	return globalLogger
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func InfoKV(msg string, keysAndValues ...any) {
	Get().Sugar().Infow(msg, keysAndValues...)
}

func DebugKV(msg string, keysAndValues ...any) {
	Get().Sugar().Debugw(msg, keysAndValues...)
}

func ErrorKV(msg string, keysAndValues ...any) {
	Get().Sugar().Errorw(msg, keysAndValues...)
}
