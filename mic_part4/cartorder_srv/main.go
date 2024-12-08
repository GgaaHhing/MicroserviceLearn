package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/lithammer/shortuuid/v4"
	"google.golang.org/grpc"
	"net"
	"testProject/mic_part4/cartorder_srv/biz"
	"testProject/mic_part4/internal"
	"testProject/mic_part4/log"
	"testProject/mic_part4/proto/goole/pb"
	"testProject/mic_part4/util"
)

func init() {
	internal.InitDB()
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

	s := grpc.NewServer()
	pb.RegisterShopCartServiceServer(s, &biz.CartOrderServer{}) // 替换为你的服务接口和服务器实例

	if err := s.Serve(lis); err != nil {
		log.Logger.Error("GRPC 部署失败: " + err.Error())
		panic(err)
	}
}
