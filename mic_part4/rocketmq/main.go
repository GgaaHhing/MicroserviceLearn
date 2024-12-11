package main

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"microserviceLearn/mic_part4/internal"
	"os"
	"strconv"
	"time"
)

var mqAddr = fmt.Sprintf("%s:%d", internal.AppConf.RocketMqConfig.Host, internal.AppConf.RocketMqConfig.Port)
var groupName = "testGroup"

// ProducerMsg 生产者的流程
func ProducerMsg(topic string) {
	/*
		1. 怎么发送半消息，也就是要生产者出现commit/rollback才会消费的消息
		2. 延时投递，让消费者自查状态
		3. 回查机制
	*/
	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{mqAddr})),
		producer.WithRetry(2),
		producer.WithGroupName(groupName),
	)
	if err != nil {
		panic(err)
	}

	err = p.Start()
	if err != nil {
		zap.S().Error("启动 rocketmq 生产者错误：", zap.Error(err))
		os.Exit(1)
	}

	for i := 0; i < 10; i++ {
		msg := &primitive.Message{
			Topic: topic,
			Body:  []byte(fmt.Sprintf("Hello Sad Mall : %d", i)),
		}

		msg.WithDelayTimeLevel(2)
		res, err := p.SendSync(context.Background(), msg)
		if err != nil {
			zap.S().Error("发送 rocketmq 消息错误："+strconv.Itoa(i), zap.Error(err))
		} else {
			zap.S().Info("发送 rocketmq 消息成功" + strconv.Itoa(i) + " \n " + res.MsgID)
		}
	}

	err = p.Shutdown()
	if err != nil {
		zap.S().Error("rocketmq 生产者 shutdown 错误：", zap.Error(err))
		os.Exit(1)
	}
}

func ConsumeMsg(topic string) {
	/*
		1. 接收下游消费者状态？？？
	*/
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{mqAddr})),
	)
	if err != nil {
		panic(err)
	}

	// 订阅主题以供消费
	err = c.Subscribe(topic, consumer.MessageSelector{},
		func(ctx context.Context, msgList ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgList {
				fmt.Println("  订阅消息：" + msg.Topic + "  " + string(msg.Body))
			}
			//这里是订阅消息，消息处理完之后，我们需要告诉外面，我们已经成功
			return consumer.ConsumeSuccess, nil
		})
	if err != nil {
		zap.S().Error("消费消息异常：" + err.Error())
	}

	err = c.Start()
	if err != nil {
		zap.S().Error("开启消费者异常：" + err.Error())
	}

	time.Sleep(60 * time.Second)

	err = c.Shutdown()
	if err != nil {
		zap.S().Error("关闭消费者异常：" + err.Error())
	}
}

func main() {
	topic := "SadMall"
	ProducerMsg(topic)
	time.Sleep(5 * time.Second)
	ConsumeMsg(topic)
}
