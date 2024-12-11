package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/hashicorp/consul/api"
	"github.com/lithammer/shortuuid/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"microserviceLearn/mic_part4/cartorder_srv/biz"
	"microserviceLearn/mic_part4/cartorder_srv/model"
	"microserviceLearn/mic_part4/internal"
	"microserviceLearn/mic_part4/log"
	"microserviceLearn/mic_part4/proto/goole/pb"
	"microserviceLearn/mic_part4/util"
	"microserviceLearn/mic_part4/util/otgrpc"

	"net"
)

func init() {
	internal.InitDB()
}

func initJaeger() opentracing.Tracer {
	conf := internal.AppConf.JaegerConfig
	jaegerAddr := fmt.Sprintf("%s:%d", conf.AgentHost, conf.AgentPort)
	cfg := config.Configuration{
		ServiceName: "sadMall",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1, // 1 表示采样所有请求
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: jaegerAddr, // Jaeger Agent 的地址和端口
			LogSpans:           true,       // 在控制台打印 spans（可选）
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	defer closer.Close()
	if err != nil {
		panic(err)
	}
	return tracer
}

func main() {
	conf := internal.AppConf
	port := util.GenRandomPort()
	srvAddress := fmt.Sprintf("%s:%d", conf.ShopCartSrvConfig.Host, port)

	// 创建一个新的 Consul 客户端
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = fmt.Sprintf("%s:%d", conf.ConsulConfig.Host, conf.ConsulConfig.Port)

	consulClient, err := api.NewClient(defaultConfig)
	if err != nil {
		log.Logger.Error("创建 consulClient失败： " + err.Error())
		panic(err)
	}

	randId := shortuuid.New()
	req := &api.AgentServiceRegistration{
		Address: conf.ShopCartSrvConfig.Host,
		Port:    port,
		Name:    conf.ShopCartSrvConfig.SrvName,
		ID:      randId,
		Tags:    conf.ShopCartSrvConfig.Tags,
	}

	err = consulClient.Agent().ServiceRegister(req)
	if err != nil {
		log.Logger.Error("GRPC 部署 consul失败：" + err.Error())
		panic(err)
	}

	// 监听并服务 gRPC 请求
	lis, err := net.Listen("tcp", srvAddress)
	if err != nil {
		log.Logger.Error(srvAddress + "监听失败：" + err.Error())
		panic(err)
	}
	fmt.Println("gRPC 正在监听 :  " + srvAddress)

	tracer := initJaeger()
	opentracing.SetGlobalTracer(tracer)
	s := grpc.NewServer(grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
	pb.RegisterShopCartServiceServer(s, &biz.CartOrderServer{}) // 替换为你的服务接口和服务器实例

	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"localhost:9876"}),
		consumer.WithGroupName("SadOrderTimeout"),
	)
	if err != nil {
		log.Logger.Error("  order 构建 consumer 失败：" + err.Error())
	}

	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"localhost:9876"}),
	)
	if err != nil {
		log.Logger.Error("  order 构建 producer 失败：" + err.Error())
	}

	err = p.Start()
	if err != nil {
		log.Logger.Error("  order 启动 producer 失败：" + err.Error())
	}

	c.Subscribe("Timeout_Order_Info", consumer.MessageSelector{},
		func(ctx context.Context, messageExt ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, message := range messageExt {
				var order *model.OrderItem
				_ = json.Unmarshal(message.Body, &order)

				var temp *model.OrderItem
				r := internal.DB.Model(&model.OrderItem{}).Where(&model.OrderItem{
					OrderNo: order.OrderNo,
				}).First(temp)
				if r.RowsAffected < 1 {
					return consumer.ConsumeSuccess, nil
				}

				if temp.Status != model.PaySuc {
					tx := internal.DB.Begin()
					order.Status = model.OrderClosed
					tx.Save(&order)
					_, err = p.SendSync(ctx, primitive.NewMessage("Sad_BackStockTopic", message.Body))
					if err != nil {
						tx.Rollback()
						log.Logger.Error("  订单超时重新归还库存失败：" + err.Error())
						return consumer.ConsumeRetryLater, nil
					}
				}

			}
			return consumer.ConsumeSuccess, nil
		},
	)

	if err := s.Serve(lis); err != nil {
		log.Logger.Error("GRPC 部署失败: " + err.Error())
		panic(err)
	}
}
