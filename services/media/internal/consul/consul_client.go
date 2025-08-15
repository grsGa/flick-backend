package consul

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
)

// ConsulClient wraps Consul API client
type ConsulClient struct {
	client *api.Client
}

// NewConsulClient creates a new Consul client
func NewConsulClient(address string) (*ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = address
	
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	
	return &ConsulClient{
		client: client,
	}, nil
}

// RegisterService registers the media service with Consul
func (c *ConsulClient) RegisterService(serviceName, serviceID, address string, port int) error {
	// Create service registration
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    []string{"media", "grpc", "microservice"},
		Port:    port,
		Address: address,
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", address, port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// Register service
	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("Service %s registered with Consul at %s:%d", serviceName, address, port)
	return nil
}

// DeregisterService deregisters the service from Consul
func (c *ConsulClient) DeregisterService(serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("Service %s deregistered from Consul", serviceID)
	return nil
}

// GetLocalIP gets the local IP address
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// ParsePort parses port string to int
func ParsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", portStr)
	}
	return port, nil
}
