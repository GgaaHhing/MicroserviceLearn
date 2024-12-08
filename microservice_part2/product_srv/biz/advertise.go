package biz

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"testProject/microservice_part2/custom_error"
	"testProject/microservice_part2/internal"
	"testProject/microservice_part2/model"
	"testProject/microservice_part2/proto/google/pb"
)

func (p ProductServer) AdvertiseLIst(ctx context.Context, empty *emptypb.Empty) (*pb.AdvertiseRes, error) {
	var advertise []model.Advertise
	var advertiseList []*pb.AdvertiseItemRes
	var advertiseRes pb.AdvertiseRes

	r := internal.DB.Find(&advertise)
	if r.RowsAffected == 0 {
		return nil, errors.New(custom_error.AdvertiseNotExits)
	}

	for _, item := range advertise {
		advertiseList = append(advertiseList, AdvertiseModel2Pb(&item))
	}

	advertiseRes.Total = int32(len(advertiseList))
	advertiseRes.ItemList = advertiseList
	return &advertiseRes, nil
}

func (p ProductServer) CreateAdvertise(ctx context.Context, req *pb.AdvertiseReq) (*pb.AdvertiseItemRes, error) {
	var advertise *model.Advertise
	advertise.Index = req.Index
	advertise.Url = req.Url
	advertise.Image = req.Image
	internal.DB.Create(&advertise)
	return AdvertiseModel2Pb(advertise), nil
}

func (p ProductServer) DeleteAdvertise(ctx context.Context, req *pb.AdvertiseReq) (*emptypb.Empty, error) {
	r := internal.DB.Delete(&model.Advertise{}, req.Id)
	if r.Error != nil {
		return &emptypb.Empty{}, errors.New(custom_error.AdvertiseNotExits)
	}
	return &emptypb.Empty{}, nil
}

func (p ProductServer) UpdateAdvertise(ctx context.Context, req *pb.AdvertiseReq) (*emptypb.Empty, error) {
	var advertise *model.Advertise
	r := internal.DB.Where(advertise, req.Id)
	if r.RowsAffected == 0 {
		return nil, errors.New(custom_error.AdvertiseNotExits)
	}
	if req.Index > 0 {
		advertise.Index = req.Index
	}
	if req.Url != "" {
		advertise.Url = req.Url
	}
	if req.Image != "" {
		advertise.Image = req.Image
	}
	internal.DB.Save(&advertise)
	return &emptypb.Empty{}, nil
}

func AdvertiseModel2Pb(advertise *model.Advertise) *pb.AdvertiseItemRes {
	p := &pb.AdvertiseItemRes{
		Index: advertise.Index,
		Image: advertise.Image,
		Url:   advertise.Url,
	}
	if advertise.ID > 0 {
		p.Id = advertise.ID
	}
	return p
}
