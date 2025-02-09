package main

import (
	"context"
	"douyin-mall/proto/ai"
	"douyin-mall/proto/auth"
	"douyin-mall/proto/cart"
	"douyin-mall/proto/order"
	"douyin-mall/proto/product"
	"douyin-mall/proto/user"
	"fmt"
	"log"
	"time"

	"douyin-mall/pkg/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	// 初始化 Consul 客户端
	registry, err := registry.NewConsulRegistry("localhost:8500")
	if err != nil {
		log.Fatalf("Failed to create consul registry: %v", err)
	}

	// 发现用户服务
	userService, err := registry.GetService("user-service")
	if err != nil {
		log.Fatalf("Failed to discover user service: %v", err)
	}

	// 使用发现的地址连接服务
	userConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", userService.Address, userService.Port),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to connect user service: %v", err)
	}
	defer userConn.Close()
	userClient := user.NewUserServiceClient(userConn)

	// 发现认证服务
	authService, err := registry.GetService("auth-service")
	if err != nil {
		log.Fatalf("Failed to discover auth service: %v", err)
	}
	authConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", authService.Address, authService.Port),
		grpc.WithInsecure(),
	)
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

	// 测试商品服务
	fmt.Println("\n=== Testing Product Service ===")

	// 发现商品服务
	productService, err := registry.GetService("product-service")
	if err != nil {
		log.Fatalf("Failed to discover product service: %v", err)
	}
	productConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", productService.Address, productService.Port),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to connect product service: %v", err)
	}
	defer productConn.Close()

	// 创建带 token 的上下文
	md := metadata.New(map[string]string{
		"authorization": loginResp.Token,
	})
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	productClient := product.NewProductCatalogServiceClient(productConn)

	// 测试获取商品列表
	productsResp, err := productClient.ListProducts(ctx, &product.ListProductsReq{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Printf("List products failed: %v", err)
	} else {
		log.Printf("List products success: %+v", productsResp)
	}

	// 测试搜索商品
	searchResp, err := productClient.SearchProducts(ctx, &product.SearchProductsReq{
		Query: "手机",
	})
	if err != nil {
		log.Printf("Search products failed: %v", err)
	} else {
		log.Printf("Search products success: %+v", searchResp.Results)
	}

	// 测试获取单个商品
	getResp, err := productClient.GetProduct(ctx, &product.GetProductReq{
		Id: 1,
	})
	if err != nil {
		log.Printf("Get product failed: %v", err)
	} else {
		log.Printf("Get product success: %+v", getResp.Product)
	}

	// 测试购物车服务
	fmt.Println("\n=== Testing Cart Service ===")
	cartService, err := registry.GetService("cart-service")
	if err != nil {
		log.Fatalf("Failed to discover cart service: %v", err)
	}
	cartConn, err := grpc.Dial(fmt.Sprintf("%s:%d", cartService.Address, cartService.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect cart service: %v", err)
	}
	cartClient := cart.NewCartServiceClient(cartConn)

	// 添加商品到购物车
	_, err = cartClient.AddItem(ctx, &cart.AddItemReq{
		UserId: 1,
		Item: &cart.CartItem{
			ProductId: 1,
			Quantity:  2,
		},
	})
	if err != nil {
		log.Printf("Add to cart failed: %v", err)
	}

	// 获取购物车
	cartResp, err := cartClient.GetCart(ctx, &cart.GetCartReq{UserId: 1})
	if err != nil {
		log.Printf("Get cart failed: %v", err)
	} else {
		log.Printf("Cart items: %+v", cartResp.Cart.Items)
	}

	// 测试订单服务
	fmt.Println("\n=== Testing Order Service ===")
	orderService, err := registry.GetService("order-service")
	if err != nil {
		log.Fatalf("Failed to discover order service: %v", err)
	}
	orderConn, err := grpc.Dial(fmt.Sprintf("%s:%d", orderService.Address, orderService.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect order service: %v", err)
	}
	orderClient := order.NewOrderServiceClient(orderConn)

	// 创建订单
	orderResp, err := orderClient.PlaceOrder(ctx, &order.PlaceOrderReq{
		UserId:       1,
		UserCurrency: "CNY",
		Email:        "test@example.com",
		Address: &order.Address{
			StreetAddress: "测试街道",
			City:          "测试城市",
			State:         "测试省份",
			Country:       "中国",
			ZipCode:       "100000",
		},
		OrderItems: []*order.OrderItem{
			{
				Item: &cart.CartItem{
					ProductId: 1,
					Quantity:  1,
				},
				Cost: 99.9,
			},
		},
	})
	if err != nil {
		log.Printf("Place order failed: %v", err)
	} else {
		log.Printf("Order created: %+v", orderResp.Order)
	}

	// 查询订单列表
	ordersResp, err := orderClient.ListOrder(ctx, &order.ListOrderReq{UserId: 1})
	if err != nil {
		log.Printf("List orders failed: %v", err)
	} else {
		log.Printf("Orders: %+v", ordersResp.Orders)
	}

	// 测试 AI 服务
	fmt.Println("\n=== Testing AI Service ===")
	aiService, err := registry.GetService("ai-service")
	if err != nil {
		log.Fatalf("Failed to discover AI service: %v", err)
	}
	aiConn, err := grpc.Dial(fmt.Sprintf("%s:%d", aiService.Address, aiService.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect AI service: %v", err)
	}
	aiClient := ai.NewAIServiceClient(aiConn)

	// 测试订单查询
	queryResp, err := aiClient.QueryOrder(ctx, &ai.QueryOrderReq{
		UserId: 1,
		Query:  "我最近的订单状态如何？",
	})
	if err != nil {
		log.Printf("AI query failed: %v", err)
	} else {
		log.Printf("AI response: %s", queryResp.Answer)
	}

	// 测试自动下单
	autoOrderResp, err := aiClient.AutoPlaceOrder(ctx, &ai.AutoOrderReq{
		UserId:      1,
		Description: "我想买一个性价比高的手机",
	})
	if err != nil {
		log.Printf("AI auto order failed: %v", err)
	} else {
		log.Printf("AI created order: %s", autoOrderResp.OrderId)
	}
}
