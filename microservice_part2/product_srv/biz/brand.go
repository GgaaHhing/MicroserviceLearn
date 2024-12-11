package biz

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"microserviceLearn/microservice_part2/custom_error"
	"microserviceLearn/microservice_part2/internal"
	"microserviceLearn/microservice_part2/model"
	"microserviceLearn/microservice_part2/proto/google/pb"
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
	r := internal.DB.Delete(&model.Brand{}, req.Id)
	if r.Error != nil {
		return nil, errors.New(custom_error.DelBrandFail)
	}
	return &emptypb.Empty{}, nil
}

func (p ProductServer) UpdateBrand(ctx context.Context, req *pb.BrandItemReq) (*emptypb.Empty, error) {
	var brand *model.Brand
	r := internal.DB.Where(&brand, req.Id)
	if r.RowsAffected == 0 {
		return nil, errors.New(custom_error.BrandNotExits)
	}
	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Logo != "" {
		brand.Logo = req.Logo
	}
	//internal.DB.Updates(brand)	//效率更高，但是只更新非零字段，而且还会有一些关于NULL的注意点，需提防
	internal.DB.Save(&brand)
	return &emptypb.Empty{}, nil
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
