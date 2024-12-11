package internal

type ShopCartSrvConfig struct {
	SrvName string   `mapstructure:"srvName" json:"srvName"`
	Host    string   `mapstructure:"host" json:"host"`
	Port    int      `mapstructure:"port" json:"port"`
	Tags    []string `mapstructure:"tags" json:"tags"`
}

type ShopCartWebConfig struct {
	SrvName string   `mapstructure:"srvName" json:"srvName"`
	Host    string   `mapstructure:"host" json:"host"`
	Port    int      `mapstructure:"port" json:"port"`
	Tags    []string `mapstructure:"tags" json:"tags"`
}

type JaegerConfig struct {
	AgentHost string `mapstructure:"agentHost" json:"agentHost"`
	AgentPort string `mapstructure:"agentPort" json:"agentPort"`
}

type RocketMqConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type AppConfig struct {
	DBConfig          DBConfig          `mapstructure:"db" json:"db"`
	RedisConfig       RedisConfig       `mapstructure:"redis" json:"redis"`
	ConsulConfig      ConsulConfig      `mapstructure:"consul" json:"consul"`
	JaegerConfig      JaegerConfig      `mapstructure:"jaeger" json:"jaeger"`
	ShopCartSrvConfig ShopCartSrvConfig `mapstructure:"shopCart_srv" json:"shopCart_srv"`
	ShopCartWebConfig ShopCartWebConfig `mapstructure:"shopCart_web" json:"shopCart_web"`
	JWTConfig         JWTConfig         `mapstructure:"jwt" json:"jwt"`
	Debug             bool              `mapstructure:"debug" json:"debug"`
	RocketMqConfig    RocketMqConfig    `mapstructure:"rocketMq" json:"rocketMq"`
}
