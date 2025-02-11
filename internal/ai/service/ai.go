package service

import (
	"context"
	"douyin-mall/pkg/aiclient"
	"douyin-mall/proto/ai"
	"douyin-mall/proto/order"
	"fmt"
	"strings"
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

	// 根据问题类型构建提示词
	var prompt string
	switch {
	case strings.Contains(req.Query, "状态"):
		prompt = fmt.Sprintf("订单状态：最近有 %d 个订单", len(orders.Orders))
	case strings.Contains(req.Query, "金额"):
		prompt = fmt.Sprintf("订单金额：共 %d 个订单", len(orders.Orders))
	case strings.Contains(req.Query, "历史"):
		prompt = fmt.Sprintf("订单历史：过去的 %d 个订单", len(orders.Orders))
	default:
		prompt = fmt.Sprintf("订单信息：总共 %d 个订单", len(orders.Orders))
	}

	// 调用 Mock AI 获取回答
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
	// 调用 Mock AI 分析商品描述
	prompt := fmt.Sprintf("推荐商品：%s", req.Description)
	suggestion, err := s.aiClient.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI suggestion: %v", err)
	}

	// TODO: 根据AI建议创建订单
	// 这里需要调用商品服务查找商品，然后创建订单

	return &ai.AutoOrderResp{
		OrderId: suggestion,
	}, nil
}
