package biz

import (
	"context"
	"errors"
	"fmt"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"testProject/microservice_part2/custom_error"
	"testProject/microservice_part2/internal"
	"testProject/microservice_part2/model"
	"testProject/microservice_part2/proto/google/pb"
)

type ProductServer struct {
	pb.ProductServiceServer
}

func ForPb(product []*model.Product) *pb.ProductRes {
	var res *pb.ProductRes
	for _, item := range product {
		res.ItemLIst = append(res.ItemLIst, ProductModel2Pb(item))
	}
	res.Total = int32(len(product))
	return res
}

func (p ProductServer) ProductList(ctx context.Context, req *pb.ProductionConditionReq) (*pb.ProductRes, error) {
	var product []*model.Product
	md := internal.DB.Model(&model.Product{})
	if req.IsPop {
		md = internal.DB.Where("is_pop=?", req.IsPop)
	}
	if req.IsNew {
		md = internal.DB.Where("is_new=?", req.IsNew)
	}
	if req.MinPrice != 0 {
		md = internal.DB.Where("min_price > ?", req.MinPrice)
	}
	if req.BrandId > 0 {
		md = internal.DB.Where("brand_id=?", req.BrandId)
	}
	if req.KeyWord != "" {
		md = internal.DB.Where("key_word like ?", "%"+req.KeyWord+"%")
	}
	if req.CategoryId > 0 {
		var category *model.Category
		if r := internal.DB.First(&category, req.CategoryId).RowsAffected; r < 1 {
			return nil, errors.New(custom_error.CategoryNotExits)
		}
		var q string
		switch category.Level {
		case 1:
			q = fmt.Sprintf("select id from category where parent_category_id in (select id from category where parent_category_id= %s)", req.CategoryId)
		case 2:
			q = fmt.Sprintf("select id from category where parent_category_id=%s)", req.CategoryId)
		case 3:
			q = fmt.Sprintf("select id from category where id=%s)", req.CategoryId)
		}
		md = internal.DB.Where(fmt.Sprintf("category_id in %s"), q)
	}
	if req.MaxPrice != 0 {
		md = internal.DB.Where("max_price>?", req.MaxPrice)
	}
	var count int64
	md.Count(&count)
	md.Joins("Category").Joins("Brand").Scopes(internal.MyPaging(req.PageNo, req.PageSize)).Find(&product)
	return ForPb(product), nil
}

// BatchGetProduct 批量获取产品
func (p ProductServer) BatchGetProduct(ctx context.Context, req *pb.BatchProductIdReq) (*pb.ProductRes, error) {
	var productList []*model.Product
	internal.DB.Find(productList, req.Ids)
	return ForPb(productList), nil

}

func (p ProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductItem) (*pb.ProductItemRes, error) {
	var category *model.Category
	var brand *model.Brand

	// TODO: 判断合法性以及是否已经存在
	if r := internal.DB.First(category, req.CategoryId).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.CategoryNotExits)
	}
	if r := internal.DB.First(brand, req.BrandId).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.BrandNotExits)
	}

	product := ProductPb2Model(req, category, brand)
	internal.DB.Save(product)
	return ProductModel2Pb(product), nil
}

func (p ProductServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductItem) (*emptypb.Empty, error) {
	if r := internal.DB.Delete(&model.Product{}, req.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.DELProductFailed)
	}
	return &emptypb.Empty{}, nil
}

func (p ProductServer) UpdateProduct(ctx context.Context, req *pb.CreateProductItem) (*emptypb.Empty, error) {
	//TODO 字段合法性判断
	var product *model.Product
	var c *model.Category
	var b *model.Brand

	if r := internal.DB.First(product, req.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.ProductNotExits)
	}
	if r := internal.DB.First(c, req.CategoryId).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.CategoryNotExits)
	}
	if r := internal.DB.First(b, req.BrandId).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.BrandNotExits)
	}
	product = ProductPb2Model(req, c, b)
	internal.DB.Save(product)
	return &emptypb.Empty{}, nil
}

// GetProductDetail 获取产品细节
func (p ProductServer) GetProductDetail(ctx context.Context, req *pb.ProductItemReq) (*pb.ProductItemRes, error) {
	var product *model.Product
	if r := internal.DB.First(product, req.Id).RowsAffected; r < 1 {
		return nil, errors.New(custom_error.ProductNotExits)
	}
	return ProductModel2Pb(product), nil
}

func (p ProductServer) mustEmbedUnimplementedProductServiceServer() {
	//TODO implement me
	panic("implement me")
}

func ProductModel2Pb(product *model.Product) *pb.ProductItemRes {
	p := &pb.ProductItemRes{
		CategoryId: product.CategoryID,
		Name:       product.Name,
		Sn:         product.SN,
		Stocks:     product.Stocks,
		SoldNum:    product.SoldNum,
		FavNum:     product.FavNum,
		Price:      product.Price,
		RealPrice:  product.RealPrice,
		ShortDesc:  product.ShortDesc,
		Desc:       product.Desc,
		Images:     product.Images,
		DescImages: product.DescImages,
		CoverImage: product.CoverImage,
		IsNew:      product.IsNew,
		IsPop:      product.IsPop,
		Selling:    product.Selling,
		AddTime:    product.CreatedAt.Unix(),
		Category:   CategoryModel2Pb(product.Category),
		Brand:      BrandModel2Pb(product.Brand),
	}
	return p
}

func ProductPb2Model(req *pb.CreateProductItem, category *model.Category,
	brand *model.Brand) *model.Product {
	p := &model.Product{
		CategoryID: req.CategoryId,
		Category:   category,
		BrandID:    req.BrandId,
		Brand:      brand,
		Selling:    req.Selling,
		ShipFree:   req.IsFree,
		IsPop:      req.IsPop,
		IsNew:      req.IsNew,
		Name:       req.Name,
		SN:         req.Sn,
		FavNum:     req.FavNum,
		SoldNum:    req.SoldNum,
		Price:      req.Price,
		RealPrice:  req.RealPrice,
		ShortDesc:  req.ShortDesc,
		Images:     req.Images,
		DescImages: req.DescImages,
		CoverImage: req.CoverImage,
	}
	return p
}
