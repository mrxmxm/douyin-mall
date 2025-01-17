package main

import (
	"douyin-mall/configs"
	"douyin-mall/internal/product/model"
	"douyin-mall/internal/product/service"
	"douyin-mall/pkg/db"
	"douyin-mall/proto/product"
	"log"
	"net"

	"google.golang.org/grpc"
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

	// 启动 gRPC 服务器
	server := grpc.NewServer()
	product.RegisterProductCatalogServiceServer(server, productService)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Product service starting on :50053")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
