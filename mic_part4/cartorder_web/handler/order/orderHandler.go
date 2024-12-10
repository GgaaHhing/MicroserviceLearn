package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"testProject/mic_part4/internal"
	"testProject/mic_part4/log"
	"testProject/mic_part4/proto/goole/pb"
	"testProject/mic_part4/util/otgrpc"
)

var orderClient pb.OrderServiceClient

func init() {
	err := initGrpcOrderClient()
	if err != nil {
		panic(err)
	}
}

func initGrpcOrderClient() error {
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
	orderClient = pb.NewOrderServiceClient(conn)

	return nil
}

func OrderListHandler(ctx *gin.Context) {
	accountId := ctx.Query("accountId")
	pageNo := ctx.Query("pageNo")
	pageSize := ctx.Query("pageSize")

}
