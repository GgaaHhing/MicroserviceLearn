package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"microserviceLearn/microservice_part2/internal"
	"microserviceLearn/microservice_part2/internal/register"
	"microserviceLearn/microservice_part2/log"
	"microserviceLearn/microservice_part2/product_web/handler"
	"microserviceLearn/microservice_part2/util"
)

var (
	consulRegistry register.ConsulRegistry
	randomId       string
)

func init() {
	//conf := internal.AppConf.ProductWebConfig
	//err := internal.Reg(conf.SrvName, conf.Host, conf.SrvName, conf.Port, conf.Tags)
	//if err != nil {
	//	panic(err)
	//}

	conf := internal.AppConf
	randomPort := util.GenRandomPort()
	if !conf.Debug {
		conf.ProductWebConfig.Port = randomPort
	}
	randomId = uuid.New().String()
	consulRegistry = register.NewConsulRegistry(conf.ProductWebConfig.Host, conf.ProductWebConfig.Port)
	err := consulRegistry.Register(conf.ProductWebConfig.SrvName, randomId, conf.ProductWebConfig.Port,
		conf.ProductWebConfig.Tags)
	if err != nil {
		log.Logger.Error("consul register err", zap.String("err", err.Error()))
	}
}

func main() {
	conf := internal.AppConf.ProductWebConfig
	ip := conf.Host
	port := util.GenRandomPort()
	if internal.AppConf.Debug {
		port = conf.Port
	}
	addr := fmt.Sprintf("%s:%d", ip, port)
	r := gin.Default()
	productGroup := r.Group("/v1/product")
	{
		productGroup.GET("/list", handler.ProductHandler)
		productGroup.POST("/add", handler.AddHandler)
		productGroup.POST("/update", handler.UpdateHandler)
		productGroup.GET("/delete/:id", handler.DeleteHandler)
		productGroup.GET("/getDetail/:id", handler.DetailHandler)
	}
	r.GET("/health", handler.HealthHandler)
	r.Run(addr)
}
