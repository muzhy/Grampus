package logger

import (
	"Grampus/internal/config"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(logConfig *config.LogConfig) *zap.Logger {
	var level zapcore.Level
	logLevel := logConfig.Level
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel // Default to info level if not specified
	}

	var writerSyncer zapcore.WriteSyncer
	if logConfig.Output == "file" {
		if _, err := os.Stat(logConfig.Path); os.IsNotExist(err) {
			os.MkdirAll(logConfig.Path, os.ModePerm)
		}
		logFilePath := logConfig.Path + "/grampus.log"
		// Use lumberjack for log rotation
		writerSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   false,
		})
	} else {
		writerSyncer = zapcore.Lock(os.Stdout)
	}
	// Set time format for the logs
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writerSyncer,
		level,
	)
	// Print Stracktrace for error level logs
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return logger
}

func GinZapLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.New().String()
		logger := zap.L().With(zap.String("request_id", requestID))
		c.Set("logger", logger) // Store logger in context for use in handlers

		c.Next() // Process request

		latency := time.Since(start)

		logger.Info("Incoming request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
