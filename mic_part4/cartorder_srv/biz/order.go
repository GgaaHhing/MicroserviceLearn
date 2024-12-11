package biz

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/lithammer/shortuuid/v4"
	"google.golang.org/grpc/codes"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"microserviceLearn/mic_part4/cartorder_srv/model"
	"microserviceLearn/mic_part4/custom_error"
	"microserviceLearn/mic_part4/internal"
	"microserviceLearn/mic_part4/log"
	"microserviceLearn/mic_part4/proto/goole/pb"
	"time"
)

type OrderListener struct {
	ID             int32
	Detail         string
	OrderNo        string
	OrderAmount    float32
	AccountId      int32
	Status         codes.Code
	Addr           string
	Receiver       string
	ReceiverMobile string
	PostCode       string
}

func (ol *OrderListener) ExecuteLocalTransaction(message *primitive.Message) primitive.LocalTransactionState {
	var orderItem *model.OrderItem
	err := json.Unmarshal(message.Body, orderItem)
	if err != nil {
		log.Logger.Error("  ExecuteLocalTransaction 反序列 orderItem 失败：" + err.Error())
		ol.Detail = "ExecuteLocalTransaction 反序列 orderItem 失败"
		return primitive.RollbackMessageState
	}

	var cartList []*model.ShopCart
	//产品和产品的购买数量
	productNumMap := make(map[int32]int32)
	// 找到用户购物车内的选中商品
	r := internal.DB.Model(&model.ShopCart{}).Where(&model.ShopCart{
		AccountId: ol.AccountId,
		Checked:   true,
	}).Find(cartList)
	if r.RowsAffected < 1 {
		ol.Detail = custom_error.OrderProductNotChecked
		ol.OrderAmount = 0
		return primitive.RollbackMessageState
	}

	var productIds []int32
	for _, item := range cartList {
		productIds = append(productIds, item.ProductId)
		productNumMap[item.ProductId] = item.Num
	}

	//获取产品的信息
	productItem, err := internal.ProductClient.BatchGetProduct(context.Background(),
		&pb.BatchProductIdReq{
			Ids: productIds,
		})
	if err != nil {
		log.Logger.Error("调用 Pb Product错误：" + err.Error())
		ol.Detail = custom_error.ProductNotFound
		ol.OrderAmount = 0
		return primitive.RollbackMessageState
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

	_, err = internal.StockClient.Sell(context.Background(), &pb.SellItem{StockItemList: stockItemList})
	if err != nil {
		log.Logger.Error("调用 Pb StockClient 失败： " + err.Error())
		ol.Detail = custom_error.StockNotEnough
		ol.OrderAmount = 0
		return primitive.RollbackMessageState
	}

	tx := internal.DB.Begin()
	//创建订单
	//orderItem := &model.OrderItem{
	//	AccountId:      ol.AccountId,
	//	OrderNo:        shortuuid.New() + time.Now().Format("2006_01_02_15:04"),
	//	Status:         "unPay",
	//	Addr:           ol.Addr,
	//	Receiver:       ol.Receiver,
	//	ReceiverMobile: ol.ReceiverMobile,
	//	PostCode:       ol.PostCode,
	//	OrderAmount:    amount,
	//}
	orderItem.Status = "unPay"
	orderItem.OrderAmount = amount

	result := tx.Save(orderItem)
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		ol.Detail = custom_error.CreateOrderError
		ol.OrderAmount = 0
		return primitive.CommitMessageState
	}

	for _, item := range orderProductList {
		item.OrderId = orderItem.ID
	}

	result = tx.CreateInBatches(orderProductList, 50)
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		ol.Detail = custom_error.CreateOrderError
		ol.OrderAmount = 0
		return primitive.CommitMessageState
	}

	//删除购物车内已购买的商品
	result = tx.Where(&model.ShopCart{
		AccountId: ol.AccountId,
		Checked:   true,
	}).Delete(&model.ShopCart{})
	if result.Error != nil || result.RowsAffected < 1 {
		tx.Rollback()
		ol.Detail = custom_error.CreateOrderError
		ol.OrderAmount = 0
		return primitive.CommitMessageState
	}

	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"localhost:9876"}))
	if err != nil {
		log.Logger.Error(" order 构建无事务producer 失败：" + err.Error())
		ol.Status = codes.Internal
		ol.Detail = " order 构建无事务producer 失败：" + err.Error()
		return primitive.CommitMessageState
	}

	err = p.Start()
	if err != nil {
		log.Logger.Error(" order 启动无事务producer 失败：" + err.Error())
		ol.Status = codes.Internal
		ol.Detail = " order 启动无事务producer 失败：" + err.Error()
		return primitive.CommitMessageState
	}
	msg := primitive.NewMessage("Timeout_Order_Info", message.Body)
	msg.WithDelayTimeLevel(6) //2min

	_, err = p.SendSync(context.Background(), msg)
	if err != nil {
		log.Logger.Error(" order 延迟发送无事务消息 失败：" + err.Error())
		ol.Status = codes.Internal
		ol.Detail = " order 延迟发送无事务消息 失败：" + err.Error()
		return primitive.CommitMessageState //库存归还
	}
	tx.Commit()
	ol.ID = orderItem.ID
	ol.OrderAmount = orderItem.OrderAmount
	ol.Status = codes.OK
	return primitive.RollbackMessageState
}

// CheckLocalTransaction ：检查本地事务
func (ol *OrderListener) CheckLocalTransaction(message *primitive.MessageExt) primitive.LocalTransactionState {
	var orderItem *model.OrderItem
	err := json.Unmarshal(message.Body, &orderItem)
	if err != nil {
		log.Logger.Error("  CheckLocalTransaction 反序列 错误：" + err.Error())
		return primitive.UnknowState
	}

	var temp *model.OrderItem
	r := internal.DB.Model(&model.OrderItem{}).Where(&model.OrderItem{
		OrderNo: orderItem.OrderNo,
	}).First(&temp).RowsAffected
	if r < 1 {
		//这里有个问题，如果在执行库存扣减中途出现问题，那么orderNO也找不到，这样会导致库存不一致，是个bug
		return primitive.CommitMessageState
	}

	return primitive.RollbackMessageState
}

// CreateOrder 创建订单 ！！没有支付，支付是另一个模块，这个只是创建
func (s CartOrderServer) CreateOrder(ctx context.Context, req *pb.OrderItemReq) (*pb.OrderItemRes, error) {
	var orderListener *OrderListener
	mqAddr := "127.0.0.1:9876"
	p, err := rocketmq.NewTransactionProducer(orderListener,
		producer.WithNameServer([]string{mqAddr}),
	)
	if err != nil {
		log.Logger.Error("  rocketmq 构建 producer失败：" + err.Error())
		return nil, err
	}

	err = p.Start()
	if err != nil {
		log.Logger.Error("  rocketmq 启动 producer失败：" + err.Error())
		return nil, err
	}

	orderItem := &model.OrderItem{
		AccountId:      req.AccountId,
		OrderNo:        shortuuid.New(),
		Addr:           req.Addr,
		Receiver:       req.Receiver,
		ReceiverMobile: req.Mobile,
		PostCode:       req.PostCode,
	}
	item, err := json.Marshal(orderItem)
	if err != nil {
		log.Logger.Error("  CreateOrder 序列化 失败：" + err.Error())
	}

	_, err = p.SendMessageInTransaction(ctx,
		//我们要向消息队列发送一个半消息，说归还库存
		primitive.NewMessage("Sad_BackStockTopic", item))
	if err != nil {
		log.Logger.Error("  rocketmq 发送带事务消息失败：" + err.Error())
		return nil, err
	}

	if orderListener.Status != codes.OK {
		return nil, errors.New(custom_error.CreateOrderError)
	}

	res := &pb.OrderItemRes{
		Id:      orderListener.ID,
		OrderNo: orderItem.OrderNo,
		Amount:  orderListener.OrderAmount,
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
		Status:     string(order.Status),
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
