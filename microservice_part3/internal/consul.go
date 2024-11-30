package internal

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func Reg(name, host, id string, port int, tags []string) error {
	// DefaultConfig返回一个默认的客户端配置
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = fmt.Sprintf("%s:%d",
		AppConf.ConsulConfig.Host, AppConf.ConsulConfig.Port)
	//构建客户端
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		return err
	}
	//构建 注册服务 需要的参数
	agentServiceRegistration := &api.AgentServiceRegistration{}
	agentServiceRegistration.Address = host
	agentServiceRegistration.ID = id
	agentServiceRegistration.Port = port
	//用于标识、分类或过滤服务实例
	agentServiceRegistration.Tags = tags
	agentServiceRegistration.Name = name
	//构造健康检查的地址
	serverAddr := fmt.Sprintf("http://%s:%d/health", host, port)
	check := api.AgentServiceCheck{
		HTTP:                           serverAddr,
		Timeout:                        "3s",
		Interval:                       "1s", //间隔
		DeregisterCriticalServiceAfter: "5s", //多少秒无法连接后 注销关键服务
	}
	agentServiceRegistration.Check = &check
	//注册服务
	return client.Agent().ServiceRegister(agentServiceRegistration)
}

func GetServiceList() error {
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = fmt.Sprintf("%s:%d",
		AppConf.ConsulConfig.Host, AppConf.ConsulConfig.Port)
	//构建客户端
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		return err
	}
	serviceList, err := client.Agent().Services()
	if err != nil {
		return err
	}
	for k, v := range serviceList {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println("------------------------")
	}
	return nil
}

// FilterService 过滤
func FilterService(filter string) error {
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = fmt.Sprintf("%s:%d",
		AppConf.ConsulConfig.Host, AppConf.ConsulConfig.Port)
	//构建客户端
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		return err
	}
	serviceList, err := client.Agent().ServicesWithFilter(filter)
	if err != nil {
		return err
	}
	for k, v := range serviceList {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println("------------------------")
	}
	return nil
}
