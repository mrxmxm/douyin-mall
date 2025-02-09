package service

import (
	"context"
	aiclient "douyin-mall/pkg/ai"
	"douyin-mall/proto/ai"
	"douyin-mall/proto/order"
	"fmt"
)

type AIService struct {
	ai.UnimplementedAIServiceServer
	orderClient order.OrderServiceClient
	aiClient    aiclient.Client
}

// NewAIService 创建AI服务实例
func NewAIService(orderClient order.OrderServiceClient) *AIService {
	// 创建 Mock AI 客户端
	mockClient := aiclient.NewMockClient()

	return &AIService{
		orderClient: orderClient,
		aiClient:    mockClient,
	}
}

// QueryOrder 处理订单查询
func (s *AIService) QueryOrder(ctx context.Context, req *ai.QueryOrderReq) (*ai.QueryOrderResp, error) {
	// 获取用户订单列表
	orders, err := s.orderClient.ListOrder(ctx, &order.ListOrderReq{
		UserId: req.UserId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %v", err)
	}

	// 构建提示词
	prompt := fmt.Sprintf("用户问题：%s\n订单信息：%v", req.Query, orders)

	// 调用大模型
	response, err := s.aiClient.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %v", err)
	}

	return &ai.QueryOrderResp{
		Answer: response,
	}, nil
}

// AutoPlaceOrder 自动下单
func (s *AIService) AutoPlaceOrder(ctx context.Context, req *ai.AutoOrderReq) (*ai.AutoOrderResp, error) {
	// 调用大模型分析商品描述
	prompt := fmt.Sprintf("根据描述推荐商品：%s", req.Description)
	suggestion, err := s.aiClient.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI suggestion: %v", err)
	}

	// TODO: 根据AI建议创建订单
	// 这里需要调用商品服务查找商品，然后创建订单

	return &ai.AutoOrderResp{
		OrderId: fmt.Sprintf("AI-ORDER-%s", suggestion), // 使用AI建议作为订单ID
	}, nil
}
