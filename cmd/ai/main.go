package main

import (
	"douyin-mall/internal/ai/service"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/ai"
	"douyin-mall/proto/order"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// 初始化 Consul 客户端
	registry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 连接订单服务
	orderConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to order service: %v", err)
	}
	orderClient := order.NewOrderServiceClient(orderConn)

	// 创建AI服务实例
	aiService := service.NewAIService(orderClient)

	// 注册服务
	serviceID := fmt.Sprintf("ai-service-%d", time.Now().Unix())
	err = registry.Register("ai-service", serviceID, "localhost", 50058)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	ai.RegisterAIServiceServer(server, aiService)

	log.Printf("AI service starting on :50058")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
