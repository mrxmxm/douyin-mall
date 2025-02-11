package main

import (
	"context"
	"douyin-mall/configs"
	"douyin-mall/internal/order/model"
	"douyin-mall/internal/order/service"
	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/order"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// 初始化数据库
	mysqlConfig := configs.NewMySQLConfig()
	mysqlClient, err := db.NewMySQLClient(mysqlConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	if err := mysqlClient.AutoMigrate(&model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	orderService := service.NewOrderService(mysqlClient)

	// 初始化 Consul 客户端
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("order-service-%d", time.Now().Unix())
	err = registry.Register("order-service", serviceID, "localhost", 50055)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	breaker := circuit.NewCircuitBreaker("order-service")

	// 启动 gRPC 服务器
	go func() {
		lis, _ := net.Listen("tcp", ":50055")
		server := grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
			}),
		)
		order.RegisterOrderServiceServer(server, orderService)
		log.Printf("Order service starting on :50055")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	h := server.Default(server.WithHostPorts(":8085"))
	h.Spin()
}
