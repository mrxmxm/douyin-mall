package model

import (
	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	gorm.Model
	UserID       uint32  `gorm:"index"`                        // 用户ID，创建索引
	OrderID      string  `gorm:"type:varchar(64);uniqueIndex"` // 订单ID，指定长度的唯一索引
	TotalAmount  float64 // 订单总金额
	Status       string  `gorm:"type:varchar(20)"`  // 订单状态：pending/paid/cancelled
	UserCurrency string  `gorm:"type:varchar(10)"`  // 用户货币类型
	Address      string  `gorm:"type:varchar(255)"` // 收货地址
	Email        string  `gorm:"type:varchar(100)"` // 用户邮箱
}

// OrderItem 订单商品模型
type OrderItem struct {
	gorm.Model
	OrderID   string  // 关联的订单ID
	ProductID uint32  // 商品ID
	Quantity  int32   // 商品数量
	UnitPrice float64 // 商品单价
}
