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

func (p ProductServer) CategoryBrandList(ctx context.Context, req *pb.PagingReq) (*pb.CateGoryBrandListRes, error) {
	// TODO: 各种逻辑判断
	var count int64
	var items []*model.ProductCategoryBrand
	var categoryBrandList []*pb.CategoryBrandRes
	var res *pb.CateGoryBrandListRes
	internal.DB.Model(&model.ProductCategoryBrand{}).Count(&count)
	res.Total = int32(count)
	internal.DB.Preload("Category").Preload("Brand").Scopes(internal.MyPaging(req.PageNo, req.PageSize)).
		Find(items)

	for _, item := range items {
		b := BrandModel2Pb(item.Brand)
		c := CategoryModel2Pb(item.Category)
		categoryBrandList = append(categoryBrandList, &pb.CategoryBrandRes{
			Brand:    b,
			Category: c,
		})
	}

	res.ItemList = categoryBrandList
	return res, nil
}

func (p ProductServer) GetCategoryBrandList(ctx context.Context, req *pb.CategoryItemReq) (*pb.BrandRes, error) {
	var brandRes *pb.BrandRes
	var category *model.Category
	var itemList []*model.ProductCategoryBrand
	var itenListRes []*pb.BrandItemRes

	if r := internal.DB.First(category, req.Id).RowsAffected; r < 1 {
		return nil, errors.New("GetCategoryBrandList: " + custom_error.CategoryBrandNotExits)
	}

	if r := internal.DB.Preload("Category").
		Where(&model.ProductCategoryBrand{CategoryID: req.Id}).Find(&itemList).
		RowsAffected; r < 1 {
		return nil, errors.New("GetCategoryBrandList: " + custom_error.CategoryBrandNotExits)
	}

	for _, item := range itemList {
		b := BrandModel2Pb(item.Brand)
		itenListRes = append(itenListRes, b)
	}
	brandRes.ItemList = itenListRes
	brandRes.Total = int32(len(itemList))
	return brandRes, nil
}

func (p ProductServer) CreateCategoryBrand(ctx context.Context, req *pb.CategoryBrandReq) (*pb.CategoryBrandRes, error) {
	var item *model.ProductCategoryBrand
	var category *model.Category
	var brand *model.Brand
	var res *pb.CategoryBrandRes

	if r := internal.DB.First(category, req.Category.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.CategoryNotExits)
	}
	if r := internal.DB.First(brand, req.Brand.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.BrandNotExits)
	}
	item.CategoryID = req.Category.Id
	item.BrandID = req.Brand.Id
	internal.DB.Save(item)
	res.Category = req.Category
	res.Brand = req.Brand
	res.Id = item.ID
	return res, nil
}

func (p ProductServer) DeleteCategoryBrand(ctx context.Context, req *pb.CategoryBrandReq) (*emptypb.Empty, error) {
	// TODO:检验字段合法性
	if err := internal.DB.Delete(&model.ProductCategoryBrand{}, req.Id).Error; err != nil {
		return &emptypb.Empty{}, errors.New("DeleteCategoryBrand" + custom_error.SystemError)
	}
	return &emptypb.Empty{}, nil
}

func (p ProductServer) UpdateCategoryBrand(ctx context.Context, req *pb.CategoryBrandReq) (*emptypb.Empty, error) {
	if r := internal.DB.First(&model.Category{}, req.Category.Id).RowsAffected; r < 1 {
		return &emptypb.Empty{}, errors.New(custom_error.CategoryNotExits)
	}
	if r := internal.DB.First(&model.Brand{}, req.Brand.Id).RowsAffected; r < 1 {
		return &emptypb.Empty{}, errors.New(custom_error.BrandNotExits)
	}

	var item *model.ProductCategoryBrand
	item.CategoryID = req.Category.Id
	item.BrandID = req.Brand.Id
	internal.DB.Save(item)
	return &emptypb.Empty{}, nil
}

func PCBModel2Pb(pcb *model.ProductCategoryBrand) *pb.CategoryBrandRes {
	p := &pb.CategoryBrandRes{
		Id:       pcb.ID,
		Category: CategoryModel2Pb(pcb.Category),
		Brand:    BrandModel2Pb(pcb.Brand),
	}
	return p
}
