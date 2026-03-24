// @title Library Management System API
// @version 1.0
// @description This is a library management system API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email dailiduzhou@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost
// @BasePath /api
// @schemes   https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/config"
	"github.com/Dailiduzhou/library_manage_sys/internal/consumers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	kafkainfra "github.com/Dailiduzhou/library_manage_sys/pkg/kafka"
	"github.com/Dailiduzhou/library_manage_sys/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/Dailiduzhou/library_manage_sys/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.ConnectDB()
	config.InitAdmin(config.DB)

	kafkaConfig := config.LoadKafkaConfig()
	saramaConfig, err := kafkainfra.NewConfig(kafkaConfig.Version)
	if err != nil {
		log.Fatalf("Kafka 配置错误: %v", err)
	}
	producer, err := kafkainfra.InitProducer(kafkaConfig.Brokers, saramaConfig)
	if err != nil {
		log.Fatalf("Kafka Producer 初始化失败: %v", err)
	}
	consumerGroup, err := kafkainfra.InitConsumerGroup(kafkaConfig.Brokers, kafkaConfig.GroupID, saramaConfig)
	if err != nil {
		log.Fatalf("Kafka ConsumerGroup 初始化失败: %v", err)
	}

	logger := middleware.InitLogger()
	defer logger.Sync()

	r := gin.Default()

	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},

		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}

	r.Use(cors.New(corsConfig))
	r.Use(middleware.GinLoggerAndMetrics(logger))
	r.Static("/uploads", "./uploads")

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if err := middleware.InitSession(r); err != nil {
		log.Printf("会话创建失败: %q", err)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	handlers, err := initializeHandlers(config.DB, producer, kafkaConfig.BorrowedTopic, kafkaConfig.ReturnedTopic)
	if err != nil {
		log.Fatalf("依赖初始化失败: %v", err)
	}

	// Register routes
	routes.RegisterUserRoutes(r, handlers.user)
	routes.RegisterBookRouters(r, handlers.book, handlers.borrow)

	log.Println("服务器启动")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	var consumerWG sync.WaitGroup
	consumerWG.Add(1)
	consumerHandler := consumers.NewBorrowEventHandler()
	consumerTopics := []string{kafkaConfig.BorrowedTopic, kafkaConfig.ReturnedTopic}
	go func() {
		defer consumerWG.Done()
		for {
			if err := consumerGroup.Consume(consumerCtx, consumerTopics, consumerHandler); err != nil {
				log.Printf("Kafka 消费失败: %v", err)
			}
			if consumerCtx.Err() != nil {
				return
			}
		}
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	consumerCancel()

	// 设置 5 秒超时，处理未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	consumerWG.Wait()
	if err := kafkainfra.Close(); err != nil {
		log.Printf("Kafka 关闭失败: %v", err)
	}

	// 关闭数据库连接
	sqlDB, err := config.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("服务器已优雅退出")
}
