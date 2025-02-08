package service

import (
	"context"
	"douyin-mall/internal/order/model"
	"douyin-mall/proto/cart"
	"douyin-mall/proto/order"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	order.UnimplementedOrderServiceServer
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) PlaceOrder(ctx context.Context, req *order.PlaceOrderReq) (*order.PlaceOrderResp, error) {
	orderID := fmt.Sprintf("ORDER-%d", time.Now().UnixNano())

	// 开启事务
	tx := s.db.Begin()

	// 创建订单
	newOrder := model.Order{
		UserID:       req.UserId,
		OrderID:      orderID,
		Status:       "pending",
		UserCurrency: req.UserCurrency,
		Address:      req.Address.String(),
		Email:        req.Email,
	}

	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建订单项
	for _, item := range req.OrderItems {
		orderItem := model.OrderItem{
			OrderID:   orderID,
			ProductID: item.Item.ProductId,
			Quantity:  item.Item.Quantity,
			UnitPrice: float64(item.Cost),
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &order.PlaceOrderResp{
		Order: &order.OrderResult{
			OrderId: orderID,
		},
	}, nil
}

func (s *OrderService) ListOrder(ctx context.Context, req *order.ListOrderReq) (*order.ListOrderResp, error) {
	var orders []model.Order
	if err := s.db.Where("user_id = ?", req.UserId).Find(&orders).Error; err != nil {
		return nil, err
	}

	result := make([]*order.Order, 0, len(orders))
	for _, o := range orders {
		// 获取订单项
		var items []model.OrderItem
		if err := s.db.Where("order_id = ?", o.OrderID).Find(&items).Error; err != nil {
			return nil, err
		}

		orderItems := make([]*order.OrderItem, 0, len(items))
		for _, item := range items {
			orderItems = append(orderItems, &order.OrderItem{
				Cost: float32(item.UnitPrice),
				Item: &cart.CartItem{
					ProductId: item.ProductID,
					Quantity:  item.Quantity,
				},
			})
		}

		result = append(result, &order.Order{
			OrderId:      o.OrderID,
			UserId:       o.UserID,
			UserCurrency: o.UserCurrency,
			OrderItems:   orderItems,
			Email:        o.Email,
			CreatedAt:    int32(o.CreatedAt.Unix()),
		})
	}

	return &order.ListOrderResp{
		Orders: result,
	}, nil
}
