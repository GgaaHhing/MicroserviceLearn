package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"net"
	"testProject/microservice_part2/internal"
	"testProject/microservice_part2/log"
	"testProject/microservice_part2/product_srv/biz"
	"testProject/microservice_part2/proto/google/pb"
	"testProject/microservice_part2/util"
)

func init() {
	internal.InitDB()
}

func main() {
	conf := internal.AppConf
	port := util.GenRandomPort()
	srvAddress := fmt.Sprintf("%s:%d", conf.ProductSrvConfig.Host, port)

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
		Address: conf.ProductSrvConfig.Host,
		Port:    port,
		Name:    conf.ProductSrvConfig.SrvName,
		ID:      randId,
		Tags:    conf.ProductSrvConfig.Tags,
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
	pb.RegisterProductServiceServer(s, &biz.ProductServer{}) // 替换为你的服务接口和服务器实例

	if err := s.Serve(lis); err != nil {
		log.Logger.Error("GRPC 部署失败: " + err.Error())
		panic(err)
	}
}
