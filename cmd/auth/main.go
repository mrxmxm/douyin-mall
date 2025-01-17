package main

import (
	"douyin-mall/internal/auth/service"
	"douyin-mall/pkg/middleware"
	"douyin-mall/proto/auth"
	"douyin-mall/proto/user"

	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// 连接用户服务
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer userConn.Close()

	userClient := user.NewUserServiceClient(userConn)

	// 创建认证服务
	authService := service.NewAuthService(userClient)

	// 创建 gRPC 服务器,添加认证中间件
	jwtSecret := []byte("your-secret-key") // 建议从配置文件读取
	authInterceptor := middleware.NewAuthInterceptor(jwtSecret)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	// 注册认证服务
	auth.RegisterAuthServiceServer(server, authService)

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Auth service starting on :50052")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
