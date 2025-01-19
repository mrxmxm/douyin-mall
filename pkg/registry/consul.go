package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

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

func (r *ConsulRegistry) Register(serviceName string, serviceID string, address string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}
	return r.client.Agent().ServiceRegister(registration)
}

func (r *ConsulRegistry) GetService(serviceName string) (*api.AgentService, error) {
	services, err := r.client.Agent().Services()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.Service == serviceName {
			return service, nil
		}
	}

	return nil, fmt.Errorf("service %s not found", serviceName)
}

func (r *ConsulRegistry) Deregister(serviceID string) error {
	return r.client.Agent().ServiceDeregister(serviceID)
}
