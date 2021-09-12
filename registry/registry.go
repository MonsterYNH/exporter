package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type RegistryConsulConfig struct {
	Address string
	Schema  string
	Kind    string
	ID      string
	Name    string
	Tags    []string
	Meta    map[string]string
}

type Registry interface {
	ServiceRegister(address string, port int, healthPath string) error
	ServiceUnRegister() error
	Services() ([]ServiceInfo, error)
}

type ServiceInfo struct {
	Kind    string
	ID      string
	Name    string
	Service string
	Tags    []string
	Port    int
	Address string
}

type ConsulRegistry struct {
	Kind      string
	ServiceID string
	Name      string
	Tags      []string
	Meta      map[string]string
	*api.Client
}

func NewConsulRegistry(config RegistryConsulConfig) (Registry, error) {
	clientConfig := api.DefaultConfig()
	clientConfig.Address = config.Address
	clientConfig.Scheme = config.Schema

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return &ConsulRegistry{
		Client:    client,
		Kind:      config.Kind,
		ServiceID: config.ID,
		Name:      config.Name,
		Tags:      config.Tags,
		Meta:      config.Meta,
	}, nil
}

func (registry *ConsulRegistry) ServiceRegister(address string, port int, healthPath string) error {
	return registry.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Kind:    api.ServiceKind(registry.Kind),
		ID:      registry.ServiceID,
		Name:    registry.Name,
		Tags:    registry.Tags,
		Address: address,
		Port:    port,
		Checks: api.AgentServiceChecks{
			{
				HTTP:     fmt.Sprintf("http://%s:%d", address, port),
				Interval: "5s",
			},
		},
	})
}

func (registry *ConsulRegistry) ServiceUnRegister() error {
	return registry.Client.Agent().ServiceDeregister(registry.ServiceID)
}

func (registry *ConsulRegistry) Services() ([]ServiceInfo, error) {
	services, err := registry.Client.Agent().Services()
	if err != nil {
		return nil, err
	}

	infos := make([]ServiceInfo, 0)
	for name, info := range services {
		infos = append(infos, ServiceInfo{
			Kind:    string(info.Kind),
			ID:      info.ID,
			Name:    name,
			Tags:    info.Tags,
			Address: info.Address,
		})
	}

	return infos, nil
}
