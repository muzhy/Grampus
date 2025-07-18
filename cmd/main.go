package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"Grampus/internal/config"
	"Grampus/internal/logger"
	"Grampus/internal/port"
	"Grampus/internal/repository"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration:", err.Error())
	}

	logger := logger.NewLogger(&config.Log)
	defer logger.Sync() // Ensure logs are flushed before exiting

	// Replace global logger with the new logger
	loggerRestore := zap.ReplaceGlobals(logger)
	defer loggerRestore()

	sugar := logger.Sugar()
	logger.Info("Init Grampus Logger successfully")

	router := gin.New()
	// 注册全局middleware
	setupGlobalMiddleware(router)

	repo := repository.NewRepository(&config.Database)
	if repo == nil {
		logger.Fatal("Failed to create repository based on the database configuration")
		return
	}
	defer repo.Close()
	// TODO 将sqliteRepo作为参数传递给其他模块的初始化函数

	// TODO 调用各个模块，由模块自身注册路由
	setupRoutter(router, repo)
	logger.Info("Init Grampus Router successfully")

	svr := &http.Server{
		Addr:    ":" + config.Server.Port,
		Handler: router,
	}

	go func() {
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	// Graceful shutdown on SIGINT or SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exiting gracefully")
}

func setupGlobalMiddleware(router *gin.Engine) {
	//	添加自定义的日志中间件，为每个请求生成唯一ID，使用zap输出日志，统一格式
	router.Use(logger.GinZapLogMiddleware())
	router.Use(gin.Recovery())
}

func setupRoutter(router *gin.Engine, repo *repository.Repository) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"version": "1.0.0",
		})
	})

	port.AddBookingRouter(router, repo.BookRepository)
}
