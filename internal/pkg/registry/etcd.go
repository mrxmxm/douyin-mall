package registry

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	client *clientv3.Client
}

func NewEtcdRegistry(endpoints []string) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdRegistry{client: client}, nil
}
