package register

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"testProject/mic_part4/internal"
)

type IRegister interface {
	Register(name, id string, port int, tags []string) error
	DeRegister(serviceId string) error
}

type ConsulRegistry struct {
	Host string
	Port int
}

func NewConsulRegistry(host string, port int) ConsulRegistry {
	return ConsulRegistry{
		Host: host,
		Port: port,
	}
}

func (cr *ConsulRegistry) Register(name, id string, port int, tags []string) error {
	defaultConfig := api.DefaultConfig()
	conf := internal.AppConf
	defaultConfig.Address = fmt.Sprintf("%s:%d", conf.ConsulConfig.Host, conf.ConsulConfig.Port)
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		zap.S().Error("构造Consul client 错误：" + err.Error())
		return err
	}
	agentServiceReg := new(api.AgentServiceRegistration)
	agentServiceReg.Name = name
	agentServiceReg.ID = id
	agentServiceReg.Port = port
	agentServiceReg.Tags = tags
	serverAddr := fmt.Sprintf("http://%s:%d/health",
		conf.ShopCartSrvConfig.Host, conf.ShopCartSrvConfig.Port)
	check := api.AgentServiceCheck{
		HTTP:                           serverAddr,
		Timeout:                        "3s",
		Interval:                       "1s",
		DeregisterCriticalServiceAfter: "5s",
	}
	agentServiceReg.Check = &check
	err = client.Agent().ServiceRegister(agentServiceReg)
	if err != nil {
		zap.S().Error("构造Consul client Agent 错误：" + err.Error())
		return err
	}
	return nil
}

func (cr *ConsulRegistry) DeRegister(serviceId string) error {
	defaultConfig := api.DefaultConfig()
	conf := internal.AppConf
	defaultConfig.Address = fmt.Sprintf("%s:%d", conf.ConsulConfig.Host, conf.ConsulConfig.Port)
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		zap.S().Error("构造Consul client 错误：" + err.Error())
		return err
	}
	return client.Agent().ServiceDeregister(serviceId)
}
