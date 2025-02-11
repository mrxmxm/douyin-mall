package main

import (
	"context"
	"douyin-mall/internal/ai/service"
	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/ai"
	"douyin-mall/proto/order"
	"fmt"
	"log"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// 初始化 Consul 客户端，用于服务注册与发现
	registry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 连接订单服务，用于查询订单信息
	orderConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to order service: %v", err)
	}
	orderClient := order.NewOrderServiceClient(orderConn)

	// 创建 AI 服务实例
	aiService := service.NewAIService(orderClient)

	// 向 Consul 注册 AI 服务
	serviceID := fmt.Sprintf("ai-service-%d", time.Now().Unix())
	err = registry.Register("ai-service", serviceID, "localhost", 50058)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	// 参数说明：100 - 每秒允许的请求数，200 - 令牌桶容量
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	// 用于服务熔断，防止服务雪崩
	breaker := circuit.NewCircuitBreaker("ai-service")

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 创建 gRPC 服务器，添加拦截器
	server := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			// 限流检查：如果请求过多，直接拒绝
			if !limiter.Allow() {
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}

			// 记录请求日志
			logger.Info("Received request",
				zap.String("method", info.FullMethod),
				zap.Any("request", req),
			)

			// 使用断路器包装处理函数
			// 如果服务出现问题，断路器会自动断开
			resp, err := breaker.Execute(func() (interface{}, error) {
				return handler(ctx, req)
			})

			// 如果请求失败，记录错误日志
			if err != nil {
				logger.Error("Request failed",
					zap.String("method", info.FullMethod),
					zap.Error(err),
				)
			}
			return resp, err
		}),
	)

	// 注册 AI 服务到 gRPC 服务器
	ai.RegisterAIServiceServer(server, aiService)

	log.Printf("AI service starting on :50058")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
