package client

import (
	"fmt"
	"log"
	"os"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceDiscovery defines the interface for service discovery.
type ServiceDiscovery interface {
	GetServiceConn(serviceName string) (*grpc.ClientConn, error)
}

// ConsulServiceDiscovery is an implementation of ServiceDiscovery using Consul.
type ConsulServiceDiscovery struct {
	client *consul.Client
}

// NewConsulServiceDiscovery creates a new ConsulServiceDiscovery.
func NewConsulServiceDiscovery() (*ConsulServiceDiscovery, error) {
	config := consul.DefaultConfig()
	consulAddr := os.Getenv("CONSUL_AGENT_ADDR")
	if consulAddr != "" {
		config.Address = consulAddr
	}

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulServiceDiscovery{client: client}, nil
}

// GetServiceConn discovers a service from Consul and returns a gRPC client connection.
func (d *ConsulServiceDiscovery) GetServiceConn(serviceName string) (*grpc.ClientConn, error) {
	services, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	// In a real-world scenario, you would implement a load balancing strategy here.
	// For simplicity, we'll just use the first available instance.
	service := services[0].Service
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)

	log.Printf("Discovered service %s at %s", serviceName, address)

	// Set max message size to 100MB for large file uploads
	maxMsgSize := 100 * 1024 * 1024 // 100MB
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial service %s: %w", serviceName, err)
	}

	return conn, nil
}
