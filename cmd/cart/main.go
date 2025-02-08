package main

import (
	"douyin-mall/configs"
	"douyin-mall/internal/cart/model"
	"douyin-mall/internal/cart/service"
	"douyin-mall/pkg/db"
	"douyin-mall/pkg/registry"
	"douyin-mall/proto/cart"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"google.golang.org/grpc"
)

func main() {
	// 初始化数据库
	mysqlConfig := configs.NewMySQLConfig()
	mysqlClient, err := db.NewMySQLClient(mysqlConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	if err := mysqlClient.AutoMigrate(&model.CartItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	cartService := service.NewCartService(mysqlClient)

	// 初始化 Consul 客户端
	consulConfig := configs.NewConsulConfig()
	registry, err := registry.NewConsulRegistry(fmt.Sprintf("%s:%d", consulConfig.Address, consulConfig.Port))
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 注册服务
	serviceID := fmt.Sprintf("cart-service-%d", time.Now().Unix())
	err = registry.Register("cart-service", serviceID, "localhost", 50054)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 启动 gRPC 服务器
	go func() {
		lis, _ := net.Listen("tcp", ":50054")
		s := grpc.NewServer()
		cart.RegisterCartServiceServer(s, cartService)
		s.Serve(lis)
	}()

	// 启动 HTTP 服务器
	h := server.Default(server.WithHostPorts(":8084"))
	h.Spin()
}
