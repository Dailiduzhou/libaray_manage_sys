package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	mu     sync.Mutex
	global *zap.Logger
)

// Init initializes the global logger and replaces zap's global logger.
func Init() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if global != nil {
		_ = global.Sync()
	}

	global = newLogger()
	zap.ReplaceGlobals(global)
	return global
}

// L returns the global logger, initializing it if needed.
func L() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if global == nil {
		global = newLogger()
		zap.ReplaceGlobals(global)
	}

	return global
}

// S returns the global sugared logger.
func S() *zap.SugaredLogger {
	return L().Sugar()
}

// Sync flushes any buffered log entries.
func Sync() error {
	return L().Sync()
}

func newLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

func Debug(args ...any) {
	S().Debug(args...)
}

func Info(args ...any) {
	S().Info(args...)
}

func Warn(args ...any) {
	S().Warn(args...)
}

func Error(args ...any) {
	S().Error(args...)
}

func Fatal(args ...any) {
	S().Fatal(args...)
}

func Debugf(format string, args ...any) {
	S().Debugf(format, args...)
}

func Infof(format string, args ...any) {
	S().Infof(format, args...)
}

func Warnf(format string, args ...any) {
	S().Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	S().Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	S().Fatalf(format, args...)
}
