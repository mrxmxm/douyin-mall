package main

import (
	"douyin-mall/configs"
	"douyin-mall/internal/payment/model"
	"douyin-mall/internal/payment/service"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/payment"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// 初始化数据库连接
	mysqlConfig := configs.NewMySQLConfig()
	mysqlClient, err := db.NewMySQLClient(mysqlConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表结构
	if err := mysqlClient.AutoMigrate(&model.Payment{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建支付服务实例
	paymentService := service.NewPaymentService(mysqlClient)

	// 注册服务到 Consul
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("payment-service-%d", time.Now().Unix())
	err = registry.Register("payment-service", serviceID, "localhost", 50056)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	payment.RegisterPaymentServiceServer(server, paymentService)

	log.Printf("Payment service starting on :50056")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
