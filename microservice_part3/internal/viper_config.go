package internal

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

var NacosConf = &NacosConfig{}
var AppConf AppConfig

// var ViperConf ViperConfig
var fileName = "microservice_part3/dev-config.yaml"

func initNacos() {
	v := viper.New()
	v.SetConfigFile(fileName)
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = v.Unmarshal(&NacosConf)
	if err != nil {
		panic(err)
	}
	fmt.Println(NacosConf.DataId)
	fmt.Println("Nacos 初始化成功.....")
}

func initFromNacos() {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: NacosConf.Host,
			Port:   NacosConf.Port,
		},
	}
	clientConfig := constant.ClientConfig{
		//dev: b40636e1-e8eb-45d7-bbfc-b2a74dc591c2
		//pro: a07dcd7a-d516-4d7e-864d-2c662239baba
		NamespaceId:         NacosConf.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug",
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(err)
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: NacosConf.DataId,
		Group:  NacosConf.Group,
	})
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(content), &AppConf)
	if err != nil {
		panic(err)
	}
	fmt.Println("Viper 初始化成功.....")

}

func init() {
	initNacos()
	initFromNacos()
	InitRedis()
}

type ViperConfig struct {
	NacosConfig NacosConfig `mapstructure:"nacos"`
}
