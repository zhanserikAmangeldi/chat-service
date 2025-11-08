package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	grpcClient "github.com/zhanserikAmangeldi/chat-service/internal/grpc"
)

type AuthMiddleware struct {
	userService *grpcClient.UserServiceClient
	jwtSecret   string
}

func NewAuthMiddleware(userService *grpcClient.UserServiceClient, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authentication token",
			})
			c.Abort()
			return
		}

		resp, err := m.userService.ValidateToken(context.Background(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to validate token",
			})
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": resp.Error,
			})
			c.Abort()
			return
		}

		c.Set("user_id", resp.UserId)
		c.Set("username", resp.Username)
		c.Set("email", resp.Email)
		c.Set("token", token)

		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	if token := c.Query("token"); token != "" {
		return token
	}

	if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
		return cookie
	}

	if token := c.GetHeader("X-Auth-Token"); token != "" {
		return token
	}

	return ""
}
