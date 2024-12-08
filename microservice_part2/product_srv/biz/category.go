package biz

import (
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"testProject/microservice_part2/custom_error"
	"testProject/microservice_part2/internal"
	"testProject/microservice_part2/model"
	"testProject/microservice_part2/proto/google/pb"
)

func (p ProductServer) GetAllCategoryList(ctx context.Context, empty *emptypb.Empty) (*pb.CategoriesRes, error) {
	var category []*model.Category
	var res *pb.CategoriesRes
	internal.DB.Where(&model.Category{Level: 1}).Preload("Subcategory.Subcategory").
		Find(&category)
	res.Total = int32(len(category))
	for _, item := range category {
		res.InfoResList = append(res.InfoResList, CategoryModel2Pb(item))
	}
	b, err := json.Marshal(res.InfoResList)
	if err != nil {
		return nil, errors.New(custom_error.MarshalCategoryError)
	}
	res.CategoryJsonFormat = string(b)
	return res, nil
}

func (p ProductServer) GetSubCategory(ctx context.Context, req *pb.CategoriesReq) (*pb.SubCategoriesRes, error) {
	var subRes *pb.SubCategoriesRes
	var preload string
	switch req.Level {
	case 1:
		preload = "Subcategory.Subcategory"
	case 2:
		preload = "Subcategory"
	case 3:
		return nil, errors.New(custom_error.SubCategoryNotExits)
	}

	var subCategory []*model.Category
	if err := internal.DB.Where(&model.Category{ParentCategoryID: req.Id}).Preload(preload).
		Find(&subCategory).Error; err != nil {
		return nil, errors.New("GetSubCategory: " + custom_error.SystemError)
	}

	for _, item := range subCategory {
		subRes.SubCategoryList = append(subRes.SubCategoryList, CategoryModel2Pb(item))
	}

	subRes.Total = int32(len(subCategory))
	b, err := json.Marshal(subRes.SubCategoryList)
	if err != nil {
		return nil, errors.New(custom_error.MarshalCategoryError)
	}
	subRes.CategoryJsonFormat = string(b)
	return subRes, nil
}

func (p ProductServer) CreateCategory(ctx context.Context, req *pb.CategoryItemReq) (*pb.CategoryItemRes, error) {
	var category *model.Category
	// TODO: 值合法性
	category.Name = req.Name
	category.Level = req.Level
	if category.Level > 1 {
		category.ParentCategoryID = req.ParentCategoryId
	}
	internal.DB.Save(&category)
	return CategoryModel2Pb(category), nil
}

func (p ProductServer) DeleteCategory(ctx context.Context, req *pb.CategoryDelReq) (*emptypb.Empty, error) {
	// 首先，我们需要确定要删除的分类及其所有子分类
	var mainCategory model.Category
	if err := internal.DB.First(&mainCategory, req.Id).Error; err != nil {
		return nil, errors.New(custom_error.CategoryNotExits)
	}

	// 删除所有子分类
	subRes, err := p.GetSubCategory(ctx, &pb.CategoriesReq{
		Id:    mainCategory.ID,
		Level: mainCategory.Level,
	})
	// 如果不等于nil，只有三种可能
	switch {
	//如果没有子类
	case err == nil:
		for _, item := range subRes.SubCategoryList {
			internal.DB.Delete(&model.Category{}, item.Id)
		}
	case errors.Is(err, errors.New(custom_error.SubCategoryNotExits)):
		break
		// 默认是nil，所以直接default
	default:
		return nil, errors.New("DEL 错误" + err.Error())
	}

	// 删除主分类
	if err := internal.DB.Delete(&mainCategory).Error; err != nil {
		return nil, errors.New(custom_error.CategoryNotExits) // 使用更具体的错误代码
	}

	return &emptypb.Empty{}, nil
}

func (p ProductServer) UpdateCategory(ctx context.Context, req *pb.CategoryItemReq) (*emptypb.Empty, error) {
	var category *model.Category
	if r := internal.DB.First(&category, req.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.CategoryNotExits)
	}

	switch {
	case req.Name != "":
		category.Name = req.Name
	case req.Level != 0:
		category.Level = req.Level
	case req.ParentCategoryId > 0:
		category.ParentCategoryID = req.ParentCategoryId
	}

	internal.DB.Save(&category)
	return &emptypb.Empty{}, nil
}

func CategoryModel2Pb(category *model.Category) *pb.CategoryItemRes {
	p := &pb.CategoryItemRes{
		Name:  category.Name,
		Level: category.Level,
	}
	if category.ParentCategoryID > 0 {
		p.ParentCategoryId = category.ParentCategoryID
	}
	return p
}
