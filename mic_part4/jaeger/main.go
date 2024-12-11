package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"log"
	"time"
)

func initJaeger(serviceName string) (opentracing.Tracer, io.Closer, error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1, // 1 表示采样所有请求
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: "127.0.0.1:6831", // Jaeger Agent 的地址和端口
			LogSpans:           true,             // 在控制台打印 spans（可选）
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(err)
	}
	return tracer, closer, nil
}

func main() {
	tracer, closer, err := initJaeger("sadMall")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer closer.Close()

	parentSpan := tracer.StartSpan("order_web")

	// 创建一个新的 span
	cartSpan := tracer.StartSpan("cart_srv", opentracing.ChildOf(parentSpan.Context()))
	cartSpan.Finish()
	time.Sleep(time.Second)

	productSpan := tracer.StartSpan("product_srv", opentracing.ChildOf(parentSpan.Context()))
	productSpan.Finish()
	time.Sleep(time.Second)

	stockSpan := tracer.StartSpan("stock_srv", opentracing.ChildOf(parentSpan.Context()))
	stockSpan.Finish()
	time.Sleep(time.Second)

	parentSpan.Finish()
	// 模拟一些工作
	fmt.Println("我的 Go 服务正在运行并被 Jaeger 追踪!")

}
