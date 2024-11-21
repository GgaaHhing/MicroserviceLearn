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

// BrandList 获取全部品牌
func (p ProductServer) BrandList(ctx context.Context, req *pb.BrandPagingReq) (*pb.BrandRes, error) {
	var brandList []model.Brand
	var brands []*pb.BrandItemRes
	var brandRes pb.BrandRes

	// 无分页功能，适合小商城
	//find := internal.DB.Find(&brandList)
	//fmt.Println(find.RowsAffected)
	//for _, item := range brandList {
	//	brands = append(brands, BrandModel2Pb(&item))
	//
	//}
	//brandRes.Total = int32(len(brandList))
	//brandRes.ItemList = brands
	//return &brandRes, nil

	// 有分页功能,缺点是：其他的业务也可能会用到，所以需要提取出分页功能
	//if req.PageNo <= 0 {
	//	req.PageNo = 1
	//}
	//var count int64
	//offset := int(req.PageSize * (req.PageNo - 1))
	//r := internal.DB.Model(&model.Brand{}).Count(&count).Offset(offset).Limit(int(req.PageSize)).Find(&brandList)
	//if r.RowsAffected == 0 {
	//	return nil, errors.New(custom_error.BrandNotExits)
	//}
	//brandRes.Total = int32(count)
	//for _, item := range brandList {
	//	brands = append(brands, BrandModel2Pb(&item))
	//}
	//brandRes.ItemList = brands
	//return &brandRes, nil

	//第三种：提取分页功能后：
	var count int64
	r := internal.DB.Model(&model.Brand{}).Count(&count).Scopes(internal.MyPaging(req.PageNo, req.PageSize)).Find(&brandList)
	if r.RowsAffected == 0 {
		return nil, errors.New(custom_error.BrandNotExits)
	}
	for _, item := range brandList {
		brands = append(brands, BrandModel2Pb(&item))
	}
	brandRes.Total = int32(count)
	brandRes.ItemList = brands
	return &brandRes, nil

}

func (p ProductServer) CreateBrand(ctx context.Context, req *pb.BrandItemReq) (*pb.BrandItemRes, error) {
	var brand model.Brand
	find := internal.DB.Find("name=?", req.Name)
	if find.RowsAffected > 0 {
		return nil, errors.New(custom_error.BrandAlreadyExits)
	}
	brand.Name = req.Name
	brand.Logo = req.Logo
	internal.DB.Save(&brand)
	return BrandModel2Pb(&brand), nil
}

func (p ProductServer) DeleteBrand(ctx context.Context, req *pb.BrandItemReq) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p ProductServer) UpdateBrand(ctx context.Context, req *pb.BrandItemReq) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func BrandModel2Pb(brand *model.Brand) *pb.BrandItemRes {
	p := &pb.BrandItemRes{
		Name: brand.Name,
		Logo: brand.Logo,
	}
	if p.Id > 0 {
		p.Id = brand.ID
	}
	return p
}
