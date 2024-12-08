package model

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int32          `gorm:"primary_key"`
	CreatedAt *time.Time     `gorm:"column:add_time"`
	UpdatedAt *time.Time     `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ShopCart struct {
	BaseModel
	AccountId int32 `gorm:"type:int;index"`
	ProductId int32 `gorm:"type:int;index"`
	Num       int32 `gorm:"int"`
	Checked   bool
}
