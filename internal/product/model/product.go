package model

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name        string  `gorm:"type:varchar(100);not null"`
	Description string  `gorm:"type:text"`
	Picture     string  `gorm:"type:varchar(255)"`
	Price       float64 `gorm:"type:decimal(10,2);not null"`
	Categories  string  `gorm:"type:varchar(255)"` // 以逗号分隔的分类列表
}
