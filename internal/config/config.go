package config

import (
	"os"
)

type Config struct {
	Port            string
	GRPCPort        string
	UserServiceAddr string
	PostgresURL     string
	RedisAddr       string
	RabbitMQURL     string
	JWTSecret       string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		GRPCPort:        getEnv("GRPC_PORT", "9090"),
		UserServiceAddr: getEnv("GRPC_USER_SERVICE", "user-service:9091"),
		PostgresURL:     getEnv("DATABASE_URL", "postgres://user:password@postgres:5432/chatapp"),
		RedisAddr:       getEnv("REDIS_ADDR", "redis:6379"),
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
		JWTSecret:       getEnv("JWT_SECRET", "your-super-secret-key"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
