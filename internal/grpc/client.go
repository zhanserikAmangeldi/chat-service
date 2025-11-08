package grpc

import (
	"context"
	"fmt"
	"github.com/zhanserikAmangeldi/chat-service/proto/user"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserServiceClient struct {
	conn   *grpc.ClientConn
	client user.UserServiceClient
}

func NewUserServiceClient(addr string) (*UserServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := user.NewUserServiceClient(conn)

	return &UserServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *UserServiceClient) ValidateToken(ctx context.Context, token string) (*user.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &user.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return resp, nil
}

func (c *UserServiceClient) GetUserInfo(ctx context.Context, userID int64) (*user.GetUserInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUserInfo(ctx, &user.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return resp, nil
}

func (c *UserServiceClient) CheckUserExists(ctx context.Context, userID int64) (*user.CheckUserExistsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CheckUserExists(ctx, &user.CheckUserExistsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check user exists: %w", err)
	}

	return resp, nil
}

func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *UserServiceClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := c.client.ValidateToken(ctx, &user.ValidateTokenRequest{
		Token: "health-check",
	})

	if err != nil {
		log.Printf("Health check warning: %v", err)
	}

	return nil
}
