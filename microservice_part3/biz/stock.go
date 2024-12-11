package biz

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm/clause"
	"microserviceLearn/microservice_part3/custom_error"
	"microserviceLearn/microservice_part3/internal"
	"microserviceLearn/microservice_part3/model"
	"microserviceLearn/microservice_part3/proto/goole/pb"
	"sync"
)

var m sync.Mutex

type StockServer struct {
	pb.StockServiceServer
}

func (s StockServer) SetStock(ctx context.Context, req *pb.ProductStockItem) (*emptypb.Empty, error) {
	var stock *model.Stock
	// TODO：字段校验
	if r := internal.DB.Where("product_id=?", req.ProductId).First(&stock).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.ParamError)
	}

	if stock.ID < 1 {
		stock.ProductId = req.ProductId
		stock.Num = req.Num
	} else {
		stock.Num += req.Num
	}

	internal.DB.Save(stock)
	return &emptypb.Empty{}, nil
}

func (s StockServer) StockDetail(ctx context.Context, req *pb.ProductStockItem) (*pb.ProductStockItem, error) {
	var stock *model.Stock
	if r := internal.DB.Where("product_id=?", req.ProductId).First(&stock).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.ParamError)
	}
	return StockModel2Pb(stock), nil
}

func (s StockServer) Sell(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	tx := internal.DB.Begin()
	m.Lock() //为了防止并发安全，加入互斥锁，但是性能差，要用分布式锁，解决这个问题
	defer m.Unlock()
	for _, item := range req.StockItemList {
		var stock *model.Stock
		if r := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=?", item.ProductId).First(&stock).RowsAffected; r < 1 {
			tx.Rollback()
			return nil, errors.New(custom_error.ProductNotFound)
		}
		if stock.Num < item.Num {
			tx.Rollback()
			return nil, errors.New(custom_error.StockNotEnough)
		}
		stock.Num -= item.Num

		tx.Where(&model.Stock{}).Select("num").Where("product_id=? and version=?",
			item.ProductId, stock.Version).
			Updates(&model.Stock{
				Num:     stock.Num,
				Version: stock.Version + 1,
			})
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

// BackStock 单单制作归还的业务逻辑，后续的因为什么归还可以在别的业务层制作
func (s StockServer) BackStock(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	tx := internal.DB.Begin()

	for _, item := range req.StockItemList {
		var stock *model.Stock
		if r := internal.DB.Where("product_id = ?", item.ProductId).First(stock).RowsAffected; r < 1 {
			tx.Rollback()
			return nil, errors.New(custom_error.ProductNotFound)
		}
		stock.Num += item.Num
		tx.Save(stock)
	}

	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (s StockServer) mustEmbedUnimplementedStockServiceServer() {}

func StockModel2Pb(stock *model.Stock) *pb.ProductStockItem {
	return &pb.ProductStockItem{
		ProductId: stock.ProductId,
		Num:       stock.Num,
	}
}
