package main

import (
	"douyin-mall/configs"
	"douyin-mall/internal/auth/service"
	"douyin-mall/pkg/middleware"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/auth"
	"douyin-mall/proto/user"

	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"

	"go.uber.org/zap"
)

func main() {
	// 连接用户服务
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userConn.Close()

	userClient := user.NewUserServiceClient(userConn)

	// 创建认证服务
	authService := service.NewAuthService(userClient)

	// 初始化 Consul 客户端
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("auth-service-%d", time.Now().Unix())
	err = registry.Register("auth-service", serviceID, "localhost", 50052)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	limiter := ratelimit.NewRateLimiter(100, 200)
	// 初始化断路器
	breaker := circuit.NewCircuitBreaker("auth-service")

	// 创建 gRPC 服务器,添加认证中间件
	jwtSecret := []byte("your-secret-key") // 建议从配置文件读取
	authInterceptor := middleware.NewAuthInterceptor(jwtSecret)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				if !limiter.Allow() {
					return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
				}
				logger.Info("Received request", zap.String("method", info.FullMethod))
				resp, err := breaker.Execute(func() (interface{}, error) {
					return handler(ctx, req)
				})
				if err != nil {
					logger.Error("Request failed", zap.Error(err))
				}
				return resp, err
			},
		),
	)

	// 注册认证服务
	auth.RegisterAuthServiceServer(server, authService)

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Auth service starting on :50052")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
