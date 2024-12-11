package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"
	"microserviceLearn/mic_part4/cartorder_web/handler"
	"microserviceLearn/mic_part4/cartorder_web/handler/cart"
	"microserviceLearn/mic_part4/cartorder_web/handler/order"
	"microserviceLearn/mic_part4/cartorder_web/middleware"
	"microserviceLearn/mic_part4/internal"
	"microserviceLearn/mic_part4/internal/register"
	"microserviceLearn/mic_part4/log"
	"microserviceLearn/mic_part4/util"
	"os"
	"os/signal"
	"syscall"
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
	cartGroup := r.Group("/v1/cartOrder").Use(middleware.Tracing())
	{
		cartGroup.GET("/list", cart.ShopCartHandler)
		cartGroup.POST("/add", cart.AddCartHandler)
		cartGroup.POST("/update", cart.UpdateCartHandler)
		cartGroup.GET("/delete", cart.DeleteCartHandler)
	}

	orderGroup := r.Group("/v1/order").Use(middleware.Tracing())
	{
		orderGroup.GET("/list", order.ListHandler)
		orderGroup.GET("/detail/:id", order.DetailHandler)
		orderGroup.POST("/add", order.CreateHandler)
	}
	r.GET("/health", handler.HealthHandler)

	go func() {
		err := r.Run(addr)
		if err != nil {
			zap.S().Panic(addr + "启动失败" + err.Error())
		} else {
			zap.S().Info(addr + "启动成功")
		}
	}()
	//优雅退出
	q := make(chan os.Signal)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
	<-q
	err := consulRegistry.DeRegister(randomId)
	if err != nil {
		zap.S().Panic("注销失败" + randomId + ":" + err.Error())
	} else {
		zap.S().Info("注销成功：" + randomId)
	}
}
