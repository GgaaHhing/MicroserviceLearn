package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"microserviceLearn/microservice_part3/biz"
	"microserviceLearn/microservice_part3/internal"
	"microserviceLearn/microservice_part3/log"
	"microserviceLearn/microservice_part3/model"
	"microserviceLearn/microservice_part3/proto/goole/pb"
	"microserviceLearn/microservice_part3/util"
	"net"
)

func init() {
	internal.InitDB()
}

func main() {
	conf := internal.AppConf
	port := util.GenRandomPort()
	srvAddress := fmt.Sprintf("%s:%d", conf.StockSrvConfig.Host, port)

	// 创建一个新的 Consul 客户端
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = fmt.Sprintf("%s:%d", conf.ConsulConfig.Host, conf.ConsulConfig.Port)

	consulClient, err := api.NewClient(defaultConfig)
	if err != nil {
		log.Logger.Error("创建 consulClient失败： " + err.Error())
		panic(err)
	}

	randId := uuid.New().String()
	req := &api.AgentServiceRegistration{
		Address: conf.StockSrvConfig.Host,
		Port:    port,
		Name:    conf.StockSrvConfig.SrvName,
		ID:      randId,
		Tags:    conf.StockSrvConfig.Tags,
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

	s := grpc.NewServer()
	pb.RegisterStockServiceServer(s, &biz.StockServer{}) // 替换为你的服务接口和服务器实例

	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"localhost:9876"}),
		consumer.WithGroupName("SadStockGroup"),
	)
	if err != nil {
		panic(err)
	}
	err = c.Subscribe("Sad_BackStockTopic", consumer.MessageSelector{},
		func(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, message := range ext {
				var order *model.Order
				err = json.Unmarshal(message.Body, order)
				if err != nil {
					log.Logger.Error("  stock Subscribe 反序列错误：" + err.Error())
					return consumer.ConsumeSuccess, nil
				}

				tx := internal.DB.Begin()
				var detail *model.StockItemDetail
				r := tx.Where(&model.StockItemDetail{
					OrderNo: order.OrderNo,
					Status:  model.HasSell,
				}).First(&detail)
				if r.RowsAffected < 1 {
					return consumer.ConsumeSuccess, nil
				}

				for _, item := range detail.DetailList {
					ret := tx.Model(&model.Stock{ProductId: item.ProductId}).
						//直接在 GORM 的查询或更新操作中嵌入自定义的 SQL 片段
						Update("num", gorm.Expr("num + ?", item.Num))

					if ret.RowsAffected < 1 {
						return consumer.ConsumeRetryLater, nil
					}
				}

				r = tx.Model(&model.StockItemDetail{}).
					Where(&model.StockItemDetail{OrderNo: order.OrderNo}).
					Update("status", model.HasBack)
				if r.RowsAffected < 1 {
					tx.Rollback()
					return consumer.ConsumeRetryLater, nil
				}
				tx.Commit()
				return consumer.ConsumeSuccess, nil
			}
			return consumer.ConsumeSuccess, nil
		},
	)
	if err != nil {
		log.Logger.Error(" stock 事务消费失败：" + err.Error())
	}

	if err := s.Serve(lis); err != nil {
		log.Logger.Error("GRPC 部署失败: " + err.Error())
		panic(err)
	}
}
