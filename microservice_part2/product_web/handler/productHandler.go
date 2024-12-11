package handler

import "C"
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"microserviceLearn/microservice/account_web/req"
	"microserviceLearn/microservice_part2/custom_error"
	"microserviceLearn/microservice_part2/internal"
	"microserviceLearn/microservice_part2/log"
	"microserviceLearn/microservice_part2/proto/google/pb"
	"net/http"
	"strconv"
)

var client pb.ProductServiceClient

func init() {
	err := initGrpcClient()
	if err != nil {
		panic(err)
	}
}

func initGrpcClient() error {
	addr := fmt.Sprintf("%s:%d", internal.AppConf.ConsulConfig.Host, internal.AppConf.ConsulConfig.Port)
	// consul://{address}/{srvName}?wait=14
	dialAddr := fmt.Sprintf("consul://%s/%s?wait=14", addr, internal.AppConf.ProductSrvConfig.SrvName)
	//grpc.Dial 已经处理了与 Consul 的交互，对于向GRPC的请求consul会自动帮助我们分配到每个启动实例上
	conn, err := grpc.Dial(dialAddr, grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"load_balancing_policy": "round_robin"}`))
	if err != nil {
		log.Logger.Info(err.Error())
	}
	defer conn.Close()
	client = pb.NewProductServiceClient(conn)

	return nil
}

func ProductHandler(c *gin.Context) {
	var condition pb.ProductionConditionReq
	//c.ShouldBindJSON(&condition)
	//list?pageNo=1&pageSize=2
	minPriceStr := c.DefaultQuery("minPrice", "0")
	minPrice, err := strconv.Atoi(minPriceStr)
	if err != nil {
		log.Logger.Error("minPrice Error")
		C.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
		return
	}
	maxPriceStr := c.DefaultQuery("maxPrice", "0")
	maxPrice, err := strconv.Atoi(maxPriceStr)
	if err != nil {
		log.Logger.Error("maxPrice Error")
		c.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
		return
	}
	condition.MinPrice = int32(minPrice)
	condition.MaxPrice = int32(maxPrice)
	categoryIdStr := c.DefaultQuery("categoryId", "0")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		log.Logger.Error("categoryId Error")
		C.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}
	condition.CategoryId = int32(categoryId)

	brandIdStr := c.DefaultQuery("brandId", "0")
	brandId, err := strconv.Atoi(brandIdStr)
	if err != nil {
		log.Logger.Error("brandId Error")
		C.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}
	condition.BrandId = int32(brandId)

	isPopStr := c.DefaultQuery("isPop", "0")
	if isPopStr == "1" {
		condition.IsPop = true
	}

	isNewStr := c.DefaultQuery("isNew", "0")
	if isNewStr == "1" {
		condition.IsNew = true
	}

	pageNoStr := c.DefaultQuery("pageNo", "0")
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		log.Logger.Error("pageNo Error")
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}
	condition.PageNo = int32(pageNo)

	pageSizeStr := c.DefaultQuery("pageSize", "0")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		log.Logger.Error("pageSize Error")
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}
	condition.PageSize = int32(pageSize)

	condition.KeyWord = c.DefaultQuery("keyWord", "0")

	list, err := client.ProductList(c, &condition)
	if err != nil {
		log.Logger.Error("Product_Web ProList: " + err.Error())
		C.JSON(http.StatusOK, gin.H{"msg": "产品列表查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "ok",
		"total": list.Total,
		"data":  list.ItemLIst,
	})
}

func AddHandler(c *gin.Context) {
	var product *req.ProductReq
	err := c.ShouldBindJSON(product)
	if err != nil {
		log.Logger.Error("ProductReq 绑定错误: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError + "解析"})
	}

	res, err := client.CreateProduct(c, ProductReq2Pb(product))
	if err != nil {
		log.Logger.Error("CreateProduct 失败: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"msg": "创建失败"})
	}
	// TODO:后面如果有关于库存价格之类的，需要查询数据库进行验证或计算
	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": res,
	})
}

// DetailHandler Param和Query区别：
// Query：用于从 URL 的查询字符串中提取参数，这些参数通常在 URL 的 ? 之后。
// Param：用于从 URL 的路径中提取参数，这些参数是 URL 路径的一部分，通过路由变量定义。
// 路由定义: router.GET("/user/:name", userHandler)
func DetailHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Error("Detail idStr Error")
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}
	res, err := client.GetProductDetail(c, &pb.ProductItemReq{
		Id: int32(id),
	})
	if err != nil {
		log.Logger.Error("获取详情错误")
		c.JSON(http.StatusOK, gin.H{"msg": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": res,
	})
}

func DeleteHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Error("DEL idStr Error")
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}

	_, err = client.DeleteProduct(c, &pb.DeleteProductItem{Id: int32(id)})
	if err != nil {
		log.Logger.Error("DEL Product 失败: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": "删除成功",
	})
}

func UpdateHandler(c *gin.Context) {
	var productReq *req.ProductReq
	err := c.ShouldBindJSON(productReq)
	if err != nil {
		log.Logger.Error("Update Product 绑定失败：: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"msg": custom_error.ParamError})
		return
	}

	_, err = client.UpdateProduct(c, ProductReq2Pb(productReq))
	if err != nil {
		log.Logger.Error("调用 GRPC UpdateProduct 失败：: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": "更新成功",
	})
}

func ProductReq2Pb(productReq *req.ProductReq) *pb.CreateProductItem {
	item := &pb.CreateProductItem{
		Name:        productReq.Name,
		Sn:          productReq.SN,
		Price:       productReq.Price,
		RealPrice:   productReq.RealPrice,
		ShortDesc:   productReq.ShortDesc,
		ProductDesc: productReq.Desc,
		Images:      productReq.Images,
		DescImages:  productReq.DescImages,
		CoverImage:  productReq.CoverImage,
		IsNew:       productReq.IsNew,
		IsPop:       productReq.IsPop,
		Selling:     productReq.Selling,
		CategoryId:  productReq.CategoryId,
		BrandId:     productReq.BrandId,
		FavNum:      productReq.FavNum,
		SoldNum:     productReq.SoldNum,
	}
	if productReq.Id > 0 {
		item.Id = productReq.Id
	}
	return item
}
