package configs

type ConsulConfig struct {
	Address string // Consul服务器地址
	Port    int    // Consul服务器端口
}

func NewConsulConfig() *ConsulConfig {
	return &ConsulConfig{
		Address: "localhost",
		Port:    8500, // Consul默认端口
	}
}
