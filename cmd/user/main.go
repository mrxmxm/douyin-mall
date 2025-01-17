package main

import (
	"douyin-mall/configs"
	"douyin-mall/internal/user/model"
	"douyin-mall/internal/user/service"
	"douyin-mall/pkg/db"
	"douyin-mall/proto/user"
	"log"
	"net"

	"github.com/cloudwego/hertz/pkg/app/server"
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
	if err := mysqlClient.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userService := service.NewUserService(mysqlClient)

	// 启动 gRPC 服务器（在新的 goroutine 中）
	go func() {
		server := grpc.NewServer()
		user.RegisterUserServiceServer(server, userService)
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("gRPC server starting on :50051")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	h := server.Default()
	h.POST("/api/user/register", userService.RegisterHTTP)
	h.POST("/api/user/login", userService.LoginHTTP)

	log.Printf("HTTP server starting on :8080")
	h.Spin()
}
