package main

import (
	"context"
	"douyin-mall/configs"
	"douyin-mall/internal/product/model"
	"douyin-mall/internal/product/service"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/product"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"douyin-mall/pkg/circuit"
	"douyin-mall/pkg/logger"
	"douyin-mall/pkg/ratelimit"

	"github.com/cloudwego/hertz/pkg/app"
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
	if err := mysqlClient.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	productService := service.NewProductService(mysqlClient)

	// 初始化 Consul 客户端
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("product-service-%d", time.Now().Unix())
	err = registry.Register("product-service", serviceID, "localhost", 50053)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 初始化限流器
	// 参数说明：100 - 每秒允许的请求数，200 - 令牌桶容量
	limiter := ratelimit.NewRateLimiter(100, 200)

	// 初始化断路器
	// 用于服务熔断，防止服务雪崩
	breaker := circuit.NewCircuitBreaker("product-service")

	// 启动 gRPC 服务器
	go func() {
		lis, _ := net.Listen("tcp", ":50053")
		server := grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				// 限流检查：如果请求过多，直接拒绝
				if !limiter.Allow() {
					return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
				}

				// 记录请求日志
				logger.Info("Received request", zap.String("method", info.FullMethod))

				// 使用断路器包装处理函数
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
		product.RegisterProductCatalogServiceServer(server, productService)
		log.Printf("Product service starting on :50053")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	h := server.Default(server.WithHostPorts(":8083"))
	h.GET("/api/products/:id", func(ctx context.Context, c *app.RequestContext) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		resp, err := productService.GetProduct(ctx, &product.GetProductReq{
			Id: uint32(id),
		})
		if err != nil {
			c.JSON(400, err)
			return
		}
		c.JSON(200, resp)
	})
	h.Spin()
}
