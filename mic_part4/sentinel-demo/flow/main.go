package main

import (
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/logging"
	"math/rand"
	"time"
)

const resource = ""

func main() {
	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()

	err := sentinel.InitWithConfig(conf)
	if err != nil {
		panic(err)
	}
	_, err = flow.LoadRules([]*flow.Rule{
		&flow.Rule{
			Resource:               resource,
			TokenCalculateStrategy: flow.Direct, //Token计算策略
			ControlBehavior:        flow.Reject, //直接拒绝     //控制行为
			Threshold:              10,          //临界点
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		panic(err)
	}

	ch := make(chan struct{})
	//测试一下
	for i := 0; i < 2; i++ {
		go func() {
			for {
				entry, blockError := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
				if blockError != nil {
					fmt.Println("流量过大，开启限流")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					fmt.Println("限流通过")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					entry.Exit()
				}
			}
		}()
	}
	<-ch

}
