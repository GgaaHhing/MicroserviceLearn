package order

import "C"
import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"microserviceLearn/mic_part4/cartorder_web/req"
	"microserviceLearn/mic_part4/custom_error"
	"microserviceLearn/mic_part4/internal"
	"microserviceLearn/mic_part4/log"
	"microserviceLearn/mic_part4/proto/goole/pb"
	"microserviceLearn/mic_part4/util/otgrpc"
	"net/http"
	"strconv"
)

var client pb.OrderServiceClient

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
	client = pb.NewOrderServiceClient(conn)

	return nil
}

func ListHandler(ctx *gin.Context) {
	aid := ctx.Query("accountId")
	accountId, err := strconv.Atoi(aid)
	if err != nil {
		log.Logger.Error("accountId 转换错误： " + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	num := ctx.Query("pageNo")
	pageNo, err := strconv.Atoi(num)
	if err != nil {
		log.Logger.Error("pageNO 转换错误： " + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	size := ctx.Query("pageSize")
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		log.Logger.Error("pageSize 转换错误： " + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	list, err := client.OrderList(context.WithValue(context.Background(), "webContent", ctx),
		&pb.OrderPagingReq{
			AccountId: int32(accountId),
			PageNo:    int32(pageNo),
			PageSize:  int32(pageSize),
		})
	if err != nil {
		log.Logger.Info("调用Order GRPC List错误：" + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": list,
	})
}

func DetailHandler(ctx *gin.Context) {
	pid := ctx.Param("id")
	id, err := strconv.Atoi(pid)
	if err != nil {
		log.Logger.Error("OrderDetail Id 转换错误： " + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	aid := ctx.Query("accountId")
	accountId, err := strconv.Atoi(aid)
	if err != nil {
		log.Logger.Error("accountId 转换错误： " + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": custom_error.ParamError,
		})
	}

	detail, err := client.OrderDetail(context.WithValue(context.Background(), "webContent", ctx),
		&pb.OrderItemReq{
			Id:        int32(id),
			AccountId: int32(accountId),
		})
	if err != nil {
		log.Logger.Info("调用Order GRPC Detail错误：" + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "获取详情失败",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": detail,
	})
}

func CreateHandler(ctx *gin.Context) {
	var orderReq req.OrderReq
	err := ctx.ShouldBindJSON(orderReq)
	if err != nil {
		log.Logger.Error("Order Create 参数绑定错误：" + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "订单创建失败",
		})
	}

	createOrder, err := client.CreateOrder(
		//这里需要携带一个名字叫做webContent的数据
		context.WithValue(context.Background(), "webContent", ctx),
		&pb.OrderItemReq{
			Id:        orderReq.Id,
			AccountId: orderReq.AccountId,
			Addr:      orderReq.Addr,
			PostCode:  orderReq.PostCode,
			Receiver:  orderReq.Receiver,
			Mobile:    orderReq.Mobile,
			PayType:   orderReq.PayType,
		})
	if err != nil {
		log.Logger.Error("Order Create GRPC 调用错误：" + err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "订单创建失败",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "",
		"data": createOrder,
	})

}
