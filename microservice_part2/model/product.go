package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type ProductCategoryBrand struct {
	BaseModel
	CategoryID int32
	BrandID    int32
	Category   *Category
	Brand      *Brand
}

type BaseModel struct {
	ID        int32 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  gorm.DeletedAt `gorm:"index"`
}

// Category 产品分类
type Category struct {
	BaseModel
	Name             string `gorm:"type:varchar(32);not null"`
	Level            int32  `gorm:"type:int"`
	ParentCategoryID int32
	ParentCategory   *Category
	//子类品类
	SubCategory []*Category `gorm:"foreignKey:ParentCategoryID;references:ID"`
}

// Brand 品牌
type Brand struct {
	BaseModel
	Name string `gorm:"type:varchar(32);not null"`
	Logo string `gorm:"type:varchar(255);not null;default:''"`
}

// Advertise 广告
type Advertise struct {
	BaseModel
	//位置
	Index int32 `gorm:"type:int;not null;default:1"`
	//图片
	Image string `gorm:"type:varchar(255);not null;"`
	//跳转地址
	Url string `gorm:"type:varchar(255);not null;"`
	//排序
	Sort int32 `gorm:"type:int;not null;default:1"`
}

// Product 产品
type Product struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;not null"`
	Category   *Category

	BrandID int32 `gorm:"type:int;not null"`
	Brand   *Brand

	//是否还在售卖
	Selling bool `gorm:"default:false"`
	//是否包邮
	ShipFree bool `gorm:"default:false"`
	//是否热卖
	IsPop bool `gorm:"default:false"`
	//是否新品
	IsNew bool `gorm:"default:false"`
	//库存
	Stocks int32 `gorm:"type:int;default:0"`

	Name string `gorm:"type:varchar(64);not null"`
	//商品编号
	SN string `gorm:"type:varchar(64);not null"`
	//有多少人收藏
	FavNum int32 `gorm:"type:int;default:0"`
	//已售卖数量
	SoldNum int32 `gorm:"type:int;default:0"`
	//定价
	Price float32 `gorm:"type:float;not null"`
	//真实价格,有点类似pdd的拼单价和单独购买价
	RealPrice float32 `gorm:"type:float;not null"`
	//简短的商品描述
	ShortDesc  string   `gorm:"type:varchar(255);not null"`
	Desc       string   `gorm:"type:varchar(255);not null"`
	Images     []string `gorm:"type:varchar(1024);not null"`
	DescImages []string `gorm:"type:varchar(1024);not null"`
	//封面
	CoverImage string `gorm:"type:varchar(255);not null"`
}

type MyList []string

// Value 当您尝试将一个 MyList 类型的值插入到数据库中时，
// 数据库驱动程序会调用这个方法来获取一个适合存储在数据库中的值（driver.Value 类型）。
// 将 MyList（即 []string）转换为 JSON 格式的字节切片
func (myList MyList) Value() (driver.Value, error) {
	return json.Marshal(myList)
}

func (myList MyList) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), &myList)
}
