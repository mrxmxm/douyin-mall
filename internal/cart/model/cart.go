package model

import (
	"gorm.io/gorm"
)

// CartItem 购物车商品模型
type CartItem struct {
	gorm.Model
	UserID    uint32 `gorm:"index"` // 用户ID，创建索引
	ProductID uint32 // 商品ID
	Quantity  int32  // 商品数量
}
