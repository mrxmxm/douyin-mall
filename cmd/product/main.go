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

	"github.com/cloudwego/hertz/pkg/app"
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

	// 启动 gRPC 服务器
	go func() {
		lis, _ := net.Listen("tcp", ":50053")
		s := grpc.NewServer()
		product.RegisterProductCatalogServiceServer(s, productService)
		s.Serve(lis)
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
