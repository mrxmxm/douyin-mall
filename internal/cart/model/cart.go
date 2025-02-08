package model

import (
	"gorm.io/gorm"
)

type CartItem struct {
	gorm.Model
	UserID    uint32 `gorm:"index"`
	ProductID uint32
	Quantity  int32
}
