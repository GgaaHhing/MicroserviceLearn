package internal

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

var RedisClient *redis.Client

func InitRedis() {
	host := AppConf.RedisConfig.Host
	port := AppConf.RedisConfig.Port
	addr := fmt.Sprintf("%s:%d", host, port)
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	ping := RedisClient.Ping(context.Background())
	fmt.Println(ping.String())
	fmt.Println("Redis DB初始化成功...")
}
