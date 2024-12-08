package biz

import (
	"context"
	"errors"
	"github.com/lithammer/shortuuid/v4"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"testProject/mic_part4/cartorder_srv/model"
	"testProject/mic_part4/custom_error"
	"testProject/mic_part4/internal"
	"testProject/mic_part4/log"
	"testProject/mic_part4/proto/goole/pb"
	"time"
)

// CreateOrder 创建订单 ！！没有支付，支付是另一个模块，这个只是创建
func (s CartOrderServer) CreateOrder(ctx context.Context, req *pb.OrderItemReq) (*pb.OrderItemRes, error) {
	/*
		1 拿到购物车内的选中产品
		2 订单总金颜 product_srv
		3 扣减库存 stock_srv
		4 把数据写到数据库里 orderItem + orderProduct表
		5 删除购物车内的已买得产品
	*/

	var cartList []*model.ShopCart
	//产品和产品的购买数量
	productNumMap := make(map[int32]int32)
	// 找到用户购物车内的选中商品
	r := internal.DB.Where(&model.ShopCart{
		AccountId: req.AccountId,
		Checked:   true,
	}).Find(cartList)
	if r.RowsAffected < 1 {
		return nil, errors.New(custom_error.OrderProductNotChecked)
	}

	var productIds []int32
	for _, item := range cartList {
		productIds = append(productIds, item.ProductId)
		productNumMap[item.ProductId] = item.Num
	}

	//获取产品的信息
	productItem, err := internal.ProductClient.BatchGetProduct(ctx, &pb.BatchProductIdReq{
		Ids: productIds,
	})
	if err != nil {
		log.Logger.Error("调用 Pb Product错误：" + err.Error())
		return nil, errors.New(custom_error.ProductNotFound)
	}

	//计算总价，总价=单价*数量
	var amount float32
	var orderProductList []*model.OrderProduct
	var stockItemList []*pb.ProductStockItem

	for _, item := range productItem.ItemLIst {
		amount += item.RealPrice * float32(productNumMap[item.Id])
		orderProduct := &model.OrderProduct{
			ProductId:   item.Id,
			ProductName: item.Name,
			CoverImage:  item.CoverImage,
			RealPrice:   item.RealPrice,
			Num:         productNumMap[item.Id],
		}
		orderProductList = append(orderProductList, orderProduct)

		stockItem := &pb.ProductStockItem{
			ProductId: item.Id,
			Num:       productNumMap[item.Id],
		}

		stockItemList = append(stockItemList, stockItem)
	}

	_, err = internal.StockClient.Sell(ctx, &pb.SellItem{StockItemList: stockItemList})
	if err != nil {
		log.Logger.Error("调用 Pb StockClient 失败： " + err.Error())
		return nil, errors.New(custom_error.StockNotEnough)
	}

	tx := internal.DB.Begin()
	//创建订单
	orderItem := &model.OrderItem{
		AccountId:      req.AccountId,
		OrderNo:        shortuuid.New() + time.Now().Format("2006_01_02_15:04"),
		Status:         "unPay",
		Addr:           req.Addr,
		Receiver:       req.Receiver,
		ReceiverMobile: req.Mobile,
		PostCode:       req.PostCode,
		OrderAmount:    amount,
	}
	result := tx.Save(orderItem)
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		return nil, errors.New(custom_error.CreateOrderError)
	}

	for _, item := range orderProductList {
		item.OrderId = orderItem.ID
	}

	result = tx.CreateInBatches(orderProductList, 50)
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		return nil, errors.New(custom_error.CreateOrderError)
	}

	//删除购物车内已购买的商品
	result = tx.Where(&model.ShopCart{
		AccountId: req.AccountId,
		Checked:   true,
	}).Delete(&model.ShopCart{})
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		return nil, errors.New(custom_error.CreateOrderError)
	}
	tx.Commit()

	res := &pb.OrderItemRes{
		Id:      orderItem.ID,
		OrderNo: orderItem.OrderNo,
		Amount:  orderItem.OrderAmount,
	}
	return res, nil
}

func (s CartOrderServer) OrderList(ctx context.Context, req *pb.OrderPagingReq) (*pb.OrderListRes, error) {
	var orderList []*model.OrderItem
	var res *pb.OrderListRes
	var count int64

	r := internal.DB.Where(&model.OrderItem{
		AccountId: req.AccountId,
	}).Count(&count)
	if r.Error != nil || r.RowsAffected < 1 {
		return nil, errors.New(custom_error.ParamError)
	}
	res.Total = int32(count)

	internal.DB.Scopes(internal.MyPaging(req.PageNo, req.PageSize)).Find(&orderList)

	for _, item := range orderList {
		res.ItemList = append(res.ItemList, OrderModel2Pb(item))
	}
	return res, nil
}

func (s CartOrderServer) OrderDetail(ctx context.Context, req *pb.OrderItemReq) (*pb.OrderItemDetailRes, error) {
	var order *model.OrderItem
	var res *pb.OrderItemDetailRes

	r := internal.DB.Where(&model.OrderItem{
		BaseModel: model.BaseModel{ID: req.Id},
		AccountId: req.AccountId,
	}).First(order)
	if r.RowsAffected < 1 {
		return nil, errors.New(custom_error.OrderNOtFount)
	}

	res.Order = OrderModel2Pb(order)

	var opList []*model.OrderProduct
	internal.DB.Where(&model.OrderProduct{
		OrderId: order.ID,
	}).Find(opList)

	for _, item := range opList {
		res.ProductList = append(res.ProductList, OrderModel2PbProduct(item))
	}

	return res, nil
}

func (s CartOrderServer) ChangeOrderStatus(ctx context.Context, status *pb.OrderStatus) (*emptypb.Empty, error) {
	r := internal.DB.Where(&model.OrderItem{
		BaseModel: model.BaseModel{ID: status.Id},
		OrderNo:   status.OrderNo,
	}).Update("status=?", status.Status)
	if r.RowsAffected < 1 {
		return nil, errors.New(custom_error.OrderNOtFount)
	}

	return &emptypb.Empty{}, nil
}

func (s CartOrderServer) mustEmbedUnimplementedOrderServiceServer() {}

func OrderModel2Pb(order *model.OrderItem) *pb.OrderItemRes {
	now := time.Now()
	res := &pb.OrderItemRes{
		Id:         order.ID,
		AccountId:  order.AccountId,
		PayType:    order.PayType,
		OrderNo:    order.OrderNo,
		PostCode:   order.PostCode,
		Amount:     order.OrderAmount,
		Addr:       order.Addr,
		Receiver:   order.Receiver,
		Mobile:     order.ReceiverMobile,
		Status:     order.Status,
		CreateTime: now.String(),
	}
	return res
}

func OrderModel2PbProduct(order *model.OrderProduct) *pb.OrderProductRes {
	res := &pb.OrderProductRes{
		Id:          order.ID,
		OrderId:     order.OrderId,
		ProductId:   order.ProductId,
		Num:         order.Num,
		ProductName: order.ProductName,
		RealPrice:   order.RealPrice,
		CoverImage:  order.CoverImage,
	}
	return res
}
