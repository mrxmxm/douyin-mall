package model

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID       uint32 `gorm:"index"`
	OrderID      string `gorm:"uniqueIndex"`
	TotalAmount  float64
	Status       string // pending, paid, cancelled
	UserCurrency string
	Address      string
	Email        string
}

type OrderItem struct {
	gorm.Model
	OrderID   string
	ProductID uint32
	Quantity  int32
	UnitPrice float64
}
