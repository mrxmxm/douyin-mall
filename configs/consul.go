package configs

type ConsulConfig struct {
	Address string
	Port    int
}

func NewConsulConfig() *ConsulConfig {
	return &ConsulConfig{
		Address: "localhost",
		Port:    8500, // Consul 默认端口
	}
}
