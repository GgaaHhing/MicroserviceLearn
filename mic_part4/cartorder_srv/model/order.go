package model

import "time"

type OrderItem struct {
	BaseModel
	AccountId      int32  `gorm:"type:int;index"`
	OrderNo        string `gorm:"type:varchar(64);index"` //订单号
	PayType        string `gorm:"type:varchar(16)"`
	Status         string `gorm:"type:varchar(16)"`
	TradeNo        string `gorm:"type:varchar(64)"` //第三方支付平台编号，为了保证双方对账
	Addr           string `gorm:"type:varchar(64)"` //地址
	Receiver       string `gorm:"type:varchar(16)"` //用户
	ReceiverMobile string `gorm:"type:varchar(11)"` //用户手机号
	PostCode       string `gorm:"type:varchar(16)"` //邮编
	OrderAmount    float32
	PayTime        *time.Time `gorm:"type:datetime"`
}

// OrderProduct 订购商品
type OrderProduct struct {
	BaseModel
	OrderId     int32   `gorm:"type:int;index"`
	ProductId   int32   `gorm:"type:int;index"`
	ProductName string  `gorm:"type:varchar(64);index"`
	CoverImage  string  `gorm:"type:varchar(128)"`
	RealPrice   float32 //真实购买价格
	Num         int32   `gorm:"type:int"` //购买数量
}
