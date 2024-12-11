package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"microserviceLearn/mic_part4/internal"
)

func Tracing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conf := internal.AppConf.JaegerConfig
		jaegerAddr := fmt.Sprintf("%s:%d", conf.AgentHost, conf.AgentPort)
		cfg := config.Configuration{
			ServiceName: "sadMall",
			Sampler: &config.SamplerConfig{
				Type:  "const",
				Param: 1, // 1 表示采样所有请求
			},
			Reporter: &config.ReporterConfig{
				LocalAgentHostPort: jaegerAddr, // Jaeger Agent 的地址和端口
				LogSpans:           true,       // 在控制台打印 spans（可选）
			},
		}

		tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
		defer closer.Close()
		if err != nil {
			panic(err)
		}
		span := tracer.StartSpan(ctx.Request.URL.Path)
		// 这里的set和 util.otgrpc 包里的client的 OpenTracingClientInterceptor 里我们加入的代码对应
		ctx.Set("tracer", tracer)
		ctx.Set("parentSpan", span)
		ctx.Next()

	}
}
