package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"net"
	"testProject/microservice_part3/biz"
	"testProject/microservice_part3/internal"
	"testProject/microservice_part3/log"
	"testProject/microservice_part3/proto/goole/pb"
	"testProject/microservice_part3/util"
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

	if err := s.Serve(lis); err != nil {
		log.Logger.Error("GRPC 部署失败: " + err.Error())
		panic(err)
	}
}
