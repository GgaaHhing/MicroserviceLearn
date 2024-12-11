package cart

import "C"
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"microserviceLearn/mic_part4/cartorder_web/req"
	"microserviceLearn/mic_part4/internal"
	"microserviceLearn/mic_part4/log"
	"microserviceLearn/mic_part4/proto/goole/pb"
	"microserviceLearn/mic_part4/util/otgrpc"
	"net/http"
	"strconv"
)

var client pb.ShopCartServiceClient

func init() {
	err := initGrpcClient()
	if err != nil {
		panic(err)
	}
}

func initGrpcClient() error {
	addr := fmt.Sprintf("%s:%d", internal.AppConf.ConsulConfig.Host, internal.AppConf.ConsulConfig.Port)
	// consul://{address}/{srvName}?wait=14
	dialAddr := fmt.Sprintf("consul://%s/%s?wait=14", addr, internal.AppConf.ShopCartWebConfig.SrvName)
	//grpc.Dial 已经处理了与 Consul 的交互，对于向GRPC的请求consul会自动帮助我们分配到每个启动实例上
	conn, err := grpc.Dial(
		dialAddr,
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"load_balancing_policy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		log.Logger.Info(err.Error())
	}
	defer conn.Close()
	client = pb.NewShopCartServiceClient(conn)

	return nil
}

func ShopCartHandler(c *gin.Context) {
	accountId := c.Query("account_id")
	id, err := strconv.Atoi(accountId)
	if err != nil {
		log.Logger.Error("ShopCart List 传入account_id错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
	}
	list, err := client.ShopCartItemList(c, &pb.AccountReq{
		AccountId: int32(id),
	})
	if err != nil {
		log.Logger.Error("ShopCart List 调用GRPC错误：" + err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": list,
	})

}

func AddCartHandler(c *gin.Context) {
	var cart req.ShopCartReq
	err := c.ShouldBindJSON(&cart)
	if err != nil {
		log.Logger.Error("ShopCart Add 参数错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
		return
	}
	if cart.Id < 1 {
		log.Logger.Error("ShopCart Add Id 错误：" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "参数错误",
		})
		return
	}
	item, err := client.AddShopCartItem(c, &pb.ShopCartReq{
		Id:        cart.Id,
		AccountId: cart.AccountId,
		ProductId: cart.ProductId,
		Num:       cart.Num,
		Checked:   cart.Checked,
	})
	if err != nil {
		log.Logger.Error("ShopCart Add 调用GRPC错误：" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "参数错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": item,
	})

}

func DeleteCartHandler(c *gin.Context) {
	aId := c.Query("account_id")
	pId := c.Query("product_id")
	accountId, err := strconv.Atoi(aId)
	if err != nil {
		log.Logger.Error("ShopCart Del accountId参数错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
		return
	}

	productId, err := strconv.Atoi(pId)
	if err != nil {
		log.Logger.Error("ShopCart Del productId参数错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
		return
	}

	_, err = client.DeleteShopCart(c, &pb.DelShopCartItem{
		AccountId: int32(accountId),
		ProductId: int32(productId),
	})
	if err != nil {
		log.Logger.Error("ShopCart Del 调用GRPC错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": "ok",
	})
}

func UpdateCartHandler(c *gin.Context) {
	var cart req.ShopCartReq
	err := c.ShouldBindJSON(&cart)
	if err != nil {
		log.Logger.Error("ShopCart Update 参数错误：" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "参数错误",
		})
		return
	}
	if cart.Id < 1 {
		log.Logger.Error("ShopCart Add Id 错误：" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "参数错误",
		})
		return
	}
	_, err = client.UpdateShopCartItem(c, &pb.ShopCartReq{
		Id:        cart.Id,
		AccountId: cart.AccountId,
		ProductId: cart.ProductId,
		Num:       cart.Num,
		Checked:   cart.Checked,
	})
	if err != nil {
		log.Logger.Error("ShopCart Add 调用GRPC错误：" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg": "参数错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": "ok",
	})
}
