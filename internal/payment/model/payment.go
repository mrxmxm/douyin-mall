package model

import (
	"gorm.io/gorm"
)

// Payment 支付记录模型
type Payment struct {
	gorm.Model
	TransactionID string  `gorm:"uniqueIndex"` // 交易ID，唯一索引
	UserID        uint32  `gorm:"index"`       // 用户ID，创建索引
	OrderID       string  `gorm:"index"`       // 订单ID，创建索引
	Amount        float64 // 支付金额
	Status        string  // 支付状态: success/failed/pending
}
