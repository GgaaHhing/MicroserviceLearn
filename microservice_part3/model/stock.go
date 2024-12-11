package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

//part3的DB中，我们有Stock表和一个StockItemDetail表

type BaseModel struct {
	ID        int32          `gorm:"primary_key"`
	CreatedAt *time.Time     `gorm:"column:add_time"`
	UpdatedAt *time.Time     `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Stock struct {
	BaseModel
	ProductId int32 `gorm:"type:int;index"`
	Num       int32 `gorm:"type:int"`
	Version   int32 `grom:"type:int"`
}

type ProductDetail struct {
	ProductId int32
	Num       int32
}

type StockItemDetail struct {
	OrderNo    string            `gorm:"type:varchar(128),index;order_no, unique"`
	Status     OrderStatus       `gorm:"type:int"`
	DetailList ProductDetailList `gorm:"type:varchar(128)"`
}

type OrderStatus int32

const (
	HasSell OrderStatus = iota + 1
	HasBack
)

type ProductDetailList []ProductDetail

// Value 从驱动程序中赋予一个可以作用于数据库的值
func (p ProductDetailList) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 从数据库驱动程序分配一个值。
func (p ProductDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &p)
}
