package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"
	"testProject/mic_part4/cartorder_web/handler"
	"testProject/mic_part4/internal"
	"testProject/mic_part4/internal/register"
	"testProject/mic_part4/log"
	"testProject/mic_part4/util"
)

var (
	consulRegistry register.ConsulRegistry
	randomId       string
)

func init() {
	conf := internal.AppConf
	randomPort := util.GenRandomPort()
	if !conf.Debug {
		conf.ShopCartWebConfig.Port = randomPort
	}
	randomId = shortuuid.New()
	consulRegistry = register.NewConsulRegistry(conf.ShopCartWebConfig.Host, conf.ShopCartWebConfig.Port)
	err := consulRegistry.Register(conf.ShopCartWebConfig.SrvName, randomId, conf.ShopCartWebConfig.Port,
		conf.ShopCartWebConfig.Tags)
	if err != nil {
		log.Logger.Error("consul register err", zap.String("err", err.Error()))
	}
}

func main() {
	conf := internal.AppConf.ShopCartWebConfig
	ip := conf.Host
	port := util.GenRandomPort()
	if internal.AppConf.Debug {
		port = conf.Port
	}
	addr := fmt.Sprintf("%s:%d", ip, port)
	r := gin.Default()
	cartGroup := r.Group("/v1/cartOrder")
	{
		cartGroup.GET("/list", handler.ShopCartHandler)
		cartGroup.POST("/add", handler.AddHandler)
		cartGroup.POST("/update", handler.UpdateHandler)
		cartGroup.GET("/delete", handler.DeleteHandler)
	}
	r.GET("/health", handler.HealthHandler)
	r.Run(addr)
}
