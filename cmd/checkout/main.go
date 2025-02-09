package main

import (
	"douyin-mall/internal/checkout/service"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/cart"
	"douyin-mall/proto/checkout"
	"douyin-mall/proto/order"
	"douyin-mall/proto/payment"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
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

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	checkout.RegisterCheckoutServiceServer(server, checkoutService)

	log.Printf("Checkout service starting on :50057")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
