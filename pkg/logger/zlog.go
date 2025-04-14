package logger

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"syscall"
)

var Logger = getLogger()

func getLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(getCurrentLogLevel())
	newLogger, _ := config.Build(
		zap.AddStacktrace(zap.DebugLevel),
		zap.AddCallerSkip(1),
	)

	return newLogger
}

func getCurrentLogLevel() zapcore.Level {
	logLevel, _ := os.LookupEnv("HICHAT_LOGGER_LEVEL")
	var level zapcore.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warning":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.WarnLevel
	}

	return level
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

func Sync() {
	err := Logger.Sync()
	if err != nil && !errors.Is(err, syscall.ENOTTY) && err.Error() != "sync /dev/stderr: invalid argument" {
		Logger.Error("zLog Sync", zap.Any("err", err))
		return
	}
}
