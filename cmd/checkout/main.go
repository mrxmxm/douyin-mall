package main

import (
	"context"
	"douyin-mall/internal/checkout/service"
	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/cart"
	"douyin-mall/proto/checkout"
	"douyin-mall/proto/order"
	"douyin-mall/proto/payment"
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
	// 初始化 Consul 客户端用于服务发现
	registry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 连接依赖的各个微服务
	cartConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	cartClient := cart.NewCartServiceClient(cartConn)

	orderConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	orderClient := order.NewOrderServiceClient(orderConn)

	paymentConn, err := grpc.Dial("localhost:50056", grpc.WithInsecure())
	paymentClient := payment.NewPaymentServiceClient(paymentConn)

	// 创建结算服务实例
	checkoutService := service.NewCheckoutService(cartClient, orderClient, paymentClient)

	// 注册服务到 Consul
	serviceID := fmt.Sprintf("checkout-service-%d", time.Now().Unix())
	err = registry.Register("checkout-service", serviceID, "localhost", 50057)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	// 参数说明：100 - 每秒允许的请求数，200 - 令牌桶容量
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	// 用于服务熔断，防止服务雪崩
	breaker := circuit.NewCircuitBreaker("checkout-service")

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50057")
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

	// 注册结算服务到 gRPC 服务器
	checkout.RegisterCheckoutServiceServer(server, checkoutService)

	log.Printf("Checkout service starting on :50057")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
