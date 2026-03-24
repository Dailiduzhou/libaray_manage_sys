package middleware

import (
	"github.com/Dailiduzhou/library_manage_sys/pkg/logger"
	"go.uber.org/zap"
)

func InitLogger() *zap.Logger {
	return logger.L()
}
