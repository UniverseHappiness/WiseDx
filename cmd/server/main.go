// Package main is the main package for the WeKnora server
// It contains the main function and the entry point for the server
//
// @title           WeKnora API
// @version         1.0
// @description     WeKnora 知识库管理系统 API 文档
// @termsOfService  http://swagger.io/terms/
//
// @contact.name   WeKnora Github
// @contact.url    https://github.com/Tencent/WeKnora
//
// @BasePath  /api/v1
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 用户登录认证：输入 Bearer {token} 格式的 JWT 令牌

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description 租户身份认证：输入 sk- 开头的 API Key
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/container"
	"github.com/Tencent/WeKnora/internal/runtime"
	"github.com/Tencent/WeKnora/internal/tracing"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// loadEnvironment 在 main 函数开始时加载环境变量
func loadEnvironment() error {
	log.Println("正在加载环境变量...")

	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
		log.Println("将继续使用系统环境变量")
	} else {
		log.Println("成功加载 .env 文件")
	}

	// 验证必需的环境变量
	requiredVars := []string{
		"DB_DRIVER", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
	}

	var missingVars []string
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("缺少必需的环境变量: %s\n请检查 .env 文件或系统环境变量",
			strings.Join(missingVars, ", "))
	}

	// 打印配置摘要
	log.Printf("数据库配置: %s://%s:%s@%s:%s/%s",
		os.Getenv("DB_DRIVER"),
		os.Getenv("DB_USER"),
		"***", // 隐藏密码
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	return nil
}

func main() {
	log.Println("Starting main function")

	// 新增：加载环境变量
	if err := loadEnvironment(); err != nil {
		log.Fatalf("环境变量加载失败: %v", err)
	}

	// Set log format with request ID
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	// 配置日志输出到文件
	var writers []io.Writer
	writers = append(writers, os.Stdout) // 保持stdout输出

	// 检查是否需要输出到文件
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writers = append(writers, file)
			log.Printf("日志将同时输出到文件: %s", logFile)
		} else {
			log.Printf("警告: 无法打开日志文件 %s: %v", logFile, err)
		}
	}

	// 设置多重写入器
	log.SetOutput(io.MultiWriter(writers...))

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Build dependency injection container
	c := container.BuildContainer(runtime.GetContainer())

	// Run application
	err := c.Invoke(func(
		cfg *config.Config,
		router *gin.Engine,
		tracer *tracing.Tracer,
		resourceCleaner interfaces.ResourceCleaner,
	) error {
		// Create context for resource cleanup
		shutdownTimeout := cfg.Server.ShutdownTimeout
		if shutdownTimeout == 0 {
			shutdownTimeout = 30 * time.Second
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cleanupCancel()

		// Register tracer cleanup function to resource cleaner
		resourceCleaner.RegisterWithName("Tracer", func() error {
			return tracer.Cleanup(cleanupCtx)
		})

		// Create HTTP server
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler: router,
		}

		ctx, done := context.WithCancel(context.Background())
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			sig := <-signals
			log.Printf("Received signal: %v, starting server shutdown...", sig)

			// Create a context with timeout for server shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Fatalf("Server forced to shutdown: %v", err)
			}

			// Clean up all registered resources
			log.Println("Cleaning up resources...")
			errs := resourceCleaner.Cleanup(cleanupCtx)
			if len(errs) > 0 {
				log.Printf("Errors occurred during resource cleanup: %v", errs)
			}

			log.Println("Server has exited")
			done()
		}()

		// Start server
		log.Printf("Server is running at %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %v", err)
		}

		// Wait for shutdown signal
		<-ctx.Done()
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
