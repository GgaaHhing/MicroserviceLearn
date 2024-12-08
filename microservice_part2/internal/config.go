package internal

type ProductSrvConfig struct {
	SrvName string   `mapstructure:"srvName" json:"srvName"`
	Host    string   `mapstructure:"host" json:"host"`
	Port    int      `mapstructure:"port" json:"port"`
	Tags    []string `mapstructure:"tags" json:"tags"`
}

type ProductWebConfig struct {
	SrvName string   `mapstructure:"srvName" json:"srvName"`
	Host    string   `mapstructure:"host" json:"host"`
	Port    int      `mapstructure:"port" json:"port"`
	Tags    []string `mapstructure:"tags" json:"tags"`
}

type AppConfig struct {
	DBConfig         DBConfig         `mapstructure:"db" json:"db"`
	RedisConfig      RedisConfig      `mapstructure:"redis" json:"redis"`
	ConsulConfig     ConsulConfig     `mapstructure:"consul" json:"consul"`
	ProductWebConfig ProductWebConfig `mapstructure:"product_web" json:"product_web"`
	ProductSrvConfig ProductSrvConfig `mapstructure:"product_srv" json:"product_srv"`
	JWTConfig        JWTConfig        `mapstructure:"jwt" json:"jwt"`
	Debug            bool             `mapstructure:"debug" json:"debug"`
}
