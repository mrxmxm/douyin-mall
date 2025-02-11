package main

import (
	"context"
	"douyin-mall/configs"
	"douyin-mall/internal/cart/model"
	"douyin-mall/internal/cart/service"
	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/cart"
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
	// 初始化 MySQL 数据库连接
	mysqlConfig := configs.NewMySQLConfig()
	mysqlClient, err := db.NewMySQLClient(mysqlConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表结构
	// 如果表不存在则创建，存在则更新表结构
	if err := mysqlClient.AutoMigrate(&model.CartItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建购物车服务实例
	cartService := service.NewCartService(mysqlClient)

	// 初始化 Consul 客户端，用于服务注册与发现
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 向 Consul 注册购物车服务
	// serviceID 需要唯一，所以使用时间戳
	serviceID := fmt.Sprintf("cart-service-%d", time.Now().Unix())
	err = registry.Register("cart-service", serviceID, "localhost", 50054)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	// 参数说明：100 - 每秒允许的请求数，200 - 令牌桶容量
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	// 用于服务熔断，防止服务雪崩
	breaker := circuit.NewCircuitBreaker("cart-service")

	// 启动 gRPC 服务器（在新的 goroutine 中）
	go func() {
		// 监听 TCP 端口
		lis, _ := net.Listen("tcp", ":50054")

		// 创建 gRPC 服务器，添加拦截器
		server := grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				// 限流检查：如果请求过多，直接拒绝
				if !limiter.Allow() {
					return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
				}

				// 记录请求日志
				logger.Info("Received request", zap.String("method", info.FullMethod))

				// 使用断路器包装处理函数
				// 如果服务出现问题，断路器会自动断开
				resp, err := breaker.Execute(func() (interface{}, error) {
					return handler(ctx, req)
				})

				// 如果请求失败，记录错误日志
				if err != nil {
					logger.Error("Request failed", zap.Error(err))
				}
				return resp, err
			}),
		)

		// 注册购物车服务到 gRPC 服务器
		cart.RegisterCartServiceServer(server, cartService)
		log.Printf("Cart service starting on :50054")

		// 启动 gRPC 服务
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器（用于健康检查等）
	h := server.Default(server.WithHostPorts(":8084"))
	h.Spin()
}
