package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64  `gorm:"primarykey"`
	Email     string `gorm:"type:varchar(100);uniqueIndex"`
	Password  string `gorm:"type:varchar(100)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
