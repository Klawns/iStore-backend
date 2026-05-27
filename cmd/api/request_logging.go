package main

import (
	"istore/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func structuredRequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path

		logger.Info("http request started",
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("user_agent", ctx.Request.UserAgent()),
		)

		ctx.Next()

		logger.Info("http request completed",
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("status", ctx.Writer.Status()),
			zap.Int("errors", len(ctx.Errors)),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}
