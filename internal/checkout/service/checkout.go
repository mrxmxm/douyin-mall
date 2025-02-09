package service

import (
	"context"
	"douyin-mall/proto/cart"
	"douyin-mall/proto/checkout"
	"douyin-mall/proto/order"
	"douyin-mall/proto/payment"
	"fmt"
)

type CheckoutService struct {
	checkout.UnimplementedCheckoutServiceServer
	cartClient    cart.CartServiceClient
	orderClient   order.OrderServiceClient
	paymentClient payment.PaymentServiceClient
}

func NewCheckoutService(
	cartClient cart.CartServiceClient,
	orderClient order.OrderServiceClient,
	paymentClient payment.PaymentServiceClient,
) *CheckoutService {
	return &CheckoutService{
		cartClient:    cartClient,
		orderClient:   orderClient,
		paymentClient: paymentClient,
	}
}

// Checkout 处理结算请求
// 主要流程:
// 1. 获取用户购物车信息
// 2. 创建订单
// 3. 调用支付服务处理支付
// 4. 清空购物车
// 5. 返回订单ID和交易ID
func (s *CheckoutService) Checkout(ctx context.Context, req *checkout.CheckoutReq) (*checkout.CheckoutResp, error) {
	// 1. 获取购物车
	cartResp, err := s.cartClient.GetCart(ctx, &cart.GetCartReq{UserId: req.UserId})
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %v", err)
	}

	// 2. 创建订单
	var orderItems []*order.OrderItem
	for _, item := range cartResp.Cart.Items {
		orderItems = append(orderItems, &order.OrderItem{
			Item: item,
			Cost: 100.0, // 这里应该从商品服务获取实际价格
		})
	}

	orderResp, err := s.orderClient.PlaceOrder(ctx, &order.PlaceOrderReq{
		UserId:       req.UserId,
		UserCurrency: "CNY",
		Email:        req.Email,
		OrderItems:   orderItems,
		Address: &order.Address{
			StreetAddress: req.Address.StreetAddress,
			City:          req.Address.City,
			State:         req.Address.State,
			Country:       req.Address.Country,
			ZipCode:       req.Address.ZipCode,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %v", err)
	}

	// 3. 处理支付
	chargeResp, err := s.paymentClient.Charge(ctx, &payment.ChargeReq{
		UserId:     req.UserId,
		OrderId:    orderResp.Order.OrderId,
		Amount:     100.0, // 这里应该计算实际金额
		CreditCard: req.CreditCard,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to charge payment: %v", err)
	}

	// 4. 清空购物车
	_, err = s.cartClient.EmptyCart(ctx, &cart.EmptyCartReq{UserId: req.UserId})
	if err != nil {
		return nil, fmt.Errorf("failed to empty cart: %v", err)
	}

	return &checkout.CheckoutResp{
		OrderId:       orderResp.Order.OrderId,
		TransactionId: chargeResp.TransactionId,
	}, nil
}
