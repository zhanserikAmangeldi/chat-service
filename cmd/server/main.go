package main

import (
	"github.com/zhanserikAmangeldi/chat-service/internal/config"
	grpcClient "github.com/zhanserikAmangeldi/chat-service/internal/grpc"
	"github.com/zhanserikAmangeldi/chat-service/internal/handlers"
	"github.com/zhanserikAmangeldi/chat-service/internal/hub"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
	"github.com/zhanserikAmangeldi/chat-service/internal/queue"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	userServiceClient, err := grpcClient.NewUserServiceClient(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal("Failed to connect to user service:", err)
	}
	defer userServiceClient.Close()

	authMiddleware := middleware.NewAuthMiddleware(userServiceClient, cfg.JWTSecret)

	h := hub.NewHub()
	go h.Run()

	publisher, err := queue.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal("Failed to create publisher:", err)
	}
	defer publisher.Close()

	router := gin.Default()

	wsHandler := handlers.NewWebSocketHandler(h, publisher)
	router.GET("/ws", authMiddleware.Authenticate(), wsHandler.HandleConnection)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Chat service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
