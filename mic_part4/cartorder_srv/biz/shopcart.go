package biz

import (
	"context"
	"errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"testProject/mic_part4/cartorder_srv/model"
	"testProject/mic_part4/custom_error"
	"testProject/mic_part4/internal"
	"testProject/mic_part4/proto/goole/pb"
)

type CartOrderServer struct {
	pb.ShopCartServiceServer
	pb.OrderServiceServer
}

func (s *CartOrderServer) ShopCartItemList(ctx context.Context, req *pb.AccountReq) (*pb.CartItemListRes, error) {
	var sc []*model.ShopCart
	var cList *pb.CartItemListRes
	r := internal.DB.Where(&model.ShopCart{AccountId: req.AccountId}).Find(&sc)
	if r.Error != nil {
		return nil, errors.New(custom_error.ParamError)
	}
	//购物车为空
	if r.RowsAffected < 1 {
		return cList, nil
	}
	for _, item := range sc {
		cList.ItemList = append(cList.ItemList, ShopCartModel2Pb(item))
	}
	return cList, nil
}

func (s *CartOrderServer) AddShopCartItem(ctx context.Context, req *pb.ShopCartReq) (*pb.CartItemRes, error) {
	var sc *model.ShopCart
	r := internal.DB.Where(&model.ShopCart{
		AccountId: req.AccountId,
		ProductId: req.ProductId,
	}).First(&sc)
	if r.RowsAffected < 1 {
		sc = ShopCartPb2Model(req)
	} else {
		sc.Num += req.Num
		sc.Checked = req.Checked
	}
	internal.DB.Save(&sc)
	return ShopCartModel2Pb(sc), nil
}

func (s *CartOrderServer) DeleteShopCart(ctx context.Context, req *pb.DelShopCartItem) (*emptypb.Empty, error) {
	r := internal.DB.Delete(&model.ShopCart{
		AccountId: req.AccountId,
		ProductId: req.ProductId,
	})
	if r.Error != nil {
		return nil, errors.New(custom_error.ShopCartNotFound)
	}
	return &emptypb.Empty{}, nil
}

func (s *CartOrderServer) UpdateShopCartItem(ctx context.Context, req *pb.ShopCartReq) (*emptypb.Empty, error) {
	var sc *model.ShopCart
	r := internal.DB.Where(&model.ShopCart{
		AccountId: req.AccountId,
		ProductId: req.ProductId,
	}).First(&sc)
	if r.RowsAffected < 1 {
		return nil, errors.New(custom_error.ShopCartNotFound)
	}
	if req.Num < 1 {
		return nil, errors.New(custom_error.ParamError)
	}
	sc.Num = req.Num
	sc.Checked = req.Checked
	return &emptypb.Empty{}, nil
}

func (s *CartOrderServer) mustEmbedUnimplementedShopCartServiceServer() {}

func ShopCartModel2Pb(sc *model.ShopCart) *pb.CartItemRes {
	c := &pb.CartItemRes{
		AccountId: sc.AccountId,
		ProductId: sc.ProductId,
		Num:       sc.Num,
		Checked:   sc.Checked,
	}
	if sc.ID > 0 {
		c.Id = sc.ID
	}
	return c
}

func ShopCartPb2Model(req *pb.ShopCartReq) *model.ShopCart {
	sc := &model.ShopCart{
		AccountId: req.AccountId,
		ProductId: req.ProductId,
		Num:       req.Num,
		Checked:   req.Checked,
	}
	if req.Id > 0 {
		sc.ID = req.Id
	}
	return sc
}
