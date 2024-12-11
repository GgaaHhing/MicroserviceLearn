package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"time"
)

type Listener struct{}

// ExecuteLocalTransaction ：执行本地事务
// Message：信息
// LocalTransactionState：本地事务状态，1表示提交，2表示回滚，3表示未知状态
func (l Listener) ExecuteLocalTransaction(message *primitive.Message) primitive.LocalTransactionState {
	return primitive.CommitMessageState
}

// CheckLocalTransaction ：检查本地事务
func (l Listener) CheckLocalTransaction(ext *primitive.MessageExt) primitive.LocalTransactionState {
	return primitive.CommitMessageState
}

func main() {
	mqAddr := "127.0.0.1:9876"
	p, err := rocketmq.NewTransactionProducer(
		Listener{},
		//set NameServer address
		producer.WithNameServer([]string{mqAddr}),
	)
	if err != nil {
		panic(err)
	}

	err = p.Start()
	if err != nil {
		panic(err)
	}

	//发送事务消息
	res, err := p.SendMessageInTransaction(
		context.Background(),
		primitive.NewMessage("SadMallTopic", []byte("hello")),
	)
	fmt.Println(res.Status)
	if err != nil {
		zap.S().Error("  发送消息失败: " + err.Error())
		panic(err)
	}
	fmt.Println("发送成功")

	time.Sleep(time.Second * 5)
	err = p.Shutdown()
	if err != nil {
		panic(err)
	}
}
