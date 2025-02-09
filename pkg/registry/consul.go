package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// ConsulRegistry Consul服务注册器
type ConsulRegistry struct {
	client *api.Client
}

func NewConsulRegistry(addr string) (*ConsulRegistry, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulRegistry{client: client}, nil
}

// Register 注册服务到Consul
func (r *ConsulRegistry) Register(serviceName string, serviceID string, address string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,   // 服务实例唯一ID
		Name:    serviceName, // 服务名称
		Address: address,     // 服务地址
		Port:    port,        // 服务端口
		Check: &api.AgentServiceCheck{ // 健康检查配置
			HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
			Interval: "10s", // 检查间隔
			Timeout:  "5s",  // 检查超时
		},
	}
	return r.client.Agent().ServiceRegister(registration)
}

// GetService 根据服务名称获取服务实例
func (r *ConsulRegistry) GetService(serviceName string) (*api.AgentService, error) {
	services, err := r.client.Agent().Services()
	if err != nil {
		return nil, err
	}

	// 查找匹配的服务
	for _, service := range services {
		if service.Service == serviceName {
			return service, nil
		}
	}
	return nil, fmt.Errorf("service %s not found", serviceName)
}

// Deregister 从Consul注销服务
func (r *ConsulRegistry) Deregister(serviceID string) error {
	return r.client.Agent().ServiceDeregister(serviceID)
}
