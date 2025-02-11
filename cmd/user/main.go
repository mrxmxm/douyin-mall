package main

import (
	"context"
	"douyin-mall/configs"
	"douyin-mall/internal/user/model"
	"douyin-mall/internal/user/service"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/user"
	"fmt"
	"log"
	"net"
	"time"

	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// 初始化数据库连接
	mysqlConfig := configs.NewMySQLConfig()
	mysqlClient, err := db.NewMySQLClient(mysqlConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	if err := mysqlClient.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userService := service.NewUserService(mysqlClient)

	// 初始化 Consul 客户端
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("user-service-%d", time.Now().Unix())
	err = registry.Register("user-service", serviceID, "localhost", 50051)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	breaker := circuit.NewCircuitBreaker("user-service")

	// 启动 gRPC 服务器（在新的 goroutine 中）
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		server := grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				// 限流检查
				if !limiter.Allow() {
					return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
				}

				// 记录请求日志
				logger.Info("Received request",
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
				)

				// 使用断路器包装处理函数
				resp, err := breaker.Execute(func() (interface{}, error) {
					return handler(ctx, req)
				})

				if err != nil {
					logger.Error("Request failed",
						zap.String("method", info.FullMethod),
						zap.Error(err),
					)
					return nil, err
				}

				return resp, nil
			}),
		)
		user.RegisterUserServiceServer(server, userService)
		log.Printf("User service starting on :50051")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	h := server.Default(server.WithHostPorts(":8081"))
	h.POST("/api/user/register", userService.RegisterHTTP)
	h.POST("/api/user/login", userService.LoginHTTP)

	log.Printf("HTTP server starting on :8080")
	h.Spin()
}
