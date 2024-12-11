package distributed_lock

import (
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"microserviceLearn/microservice_part3/internal"
	"time"
)

func RedisLock() {
	conf := internal.AppConf.RedisConfig
	redisAddr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	//返回基于 Go-redis 的池的实现。
	pool := goredis.NewPool(client)

	rs := redsync.New(pool)
	mutexName := "product@1"
	mutex := rs.NewMutex(mutexName)

	fmt.Println(" Lock().........")
	err := mutex.Lock()
	if err != nil {
		panic(err)
	}

	fmt.Println("  Get Lock().........")
	time.Sleep(time.Second)

	fmt.Println("  UnLock().........")

	ok, err := mutex.Unlock()
	if !ok || err != nil {
		panic(err)
	}

}
