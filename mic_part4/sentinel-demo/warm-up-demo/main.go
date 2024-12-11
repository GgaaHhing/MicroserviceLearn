package main

import (
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"math/rand/v2"
	"time"
)

const resource = ""

func main() {
	err := sentinel.InitDefault()
	if err != nil {
		panic(err)
	}
	_, err = flow.LoadRules([]*flow.Rule{
		&flow.Rule{
			Resource:               resource,
			TokenCalculateStrategy: flow.WarmUp,     //Token计算策略
			ControlBehavior:        flow.Throttling, //匀速通过，防止突发大流量压垮服务器
			Threshold:              1000,            //临界点
			WarmUpPeriodSec:        30,
		},
	})
	if err != nil {
		panic(err)
	}

	var all int
	var block int
	var through int
	ch := make(chan struct{})
	//测试一下
	for i := 0; i < 10; i++ {
		go func() {
			for {
				all++
				entry, blockError := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
				if blockError != nil {
					block++
					fmt.Println("流量过大，开启限流")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					through++
					fmt.Println("限流通过")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					entry.Exit()
				}
			}
		}()
	}
	<-ch

}
