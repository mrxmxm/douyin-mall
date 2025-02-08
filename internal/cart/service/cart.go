package service

import (
	"context"
	"douyin-mall/internal/cart/model"
	"douyin-mall/proto/cart"

	"gorm.io/gorm"
)

type CartService struct {
	cart.UnimplementedCartServiceServer
	db *gorm.DB
}

func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

func (s *CartService) AddItem(ctx context.Context, req *cart.AddItemReq) (*cart.AddItemResp, error) {
	var item model.CartItem

	// 检查是否已存在
	result := s.db.Where("user_id = ? AND product_id = ?", req.UserId, req.Item.ProductId).First(&item)
	if result.Error == nil {
		// 更新数量
		item.Quantity += req.Item.Quantity
		if err := s.db.Save(&item).Error; err != nil {
			return nil, err
		}
	} else {
		// 创建新项
		item = model.CartItem{
			UserID:    req.UserId,
			ProductID: req.Item.ProductId,
			Quantity:  req.Item.Quantity,
		}
		if err := s.db.Create(&item).Error; err != nil {
			return nil, err
		}
	}

	return &cart.AddItemResp{}, nil
}

func (s *CartService) GetCart(ctx context.Context, req *cart.GetCartReq) (*cart.GetCartResp, error) {
	var items []model.CartItem
	if err := s.db.Where("user_id = ?", req.UserId).Find(&items).Error; err != nil {
		return nil, err
	}

	cartItems := make([]*cart.CartItem, 0, len(items))
	for _, item := range items {
		cartItems = append(cartItems, &cart.CartItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &cart.GetCartResp{
		Cart: &cart.Cart{
			UserId: req.UserId,
			Items:  cartItems,
		},
	}, nil
}

func (s *CartService) EmptyCart(ctx context.Context, req *cart.EmptyCartReq) (*cart.EmptyCartResp, error) {
	if err := s.db.Where("user_id = ?", req.UserId).Delete(&model.CartItem{}).Error; err != nil {
		return nil, err
	}
	return &cart.EmptyCartResp{}, nil
}
