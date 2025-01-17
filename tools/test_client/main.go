package main

import (
	"context"
	"douyin-mall/proto/auth"
	"douyin-mall/proto/user"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// 连接用户服务
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect user service: %v", err)
	}
	defer userConn.Close()
	userClient := user.NewUserServiceClient(userConn)

	// 连接认证服务
	authConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect auth service: %v", err)
	}
	defer authConn.Close()
	authClient := auth.NewAuthServiceClient(authConn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 测试注册新用户
	email := fmt.Sprintf("test%d@example.com", time.Now().Unix())
	password := "test123"
	registerResp, err := userClient.Register(ctx, &user.RegisterRequest{
		Email:           email,
		Password:        password,
		ConfirmPassword: password,
	})
	if err != nil {
		log.Printf("Register failed: %v", err)
	} else {
		log.Printf("Register success: %+v", registerResp)
	}

	// 等待一下确保数据已写入
	time.Sleep(time.Second)

	// 测试登录新注册的用户
	loginResp, err := authClient.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		log.Printf("Login success, token: %s", loginResp.Token)
	}
}
