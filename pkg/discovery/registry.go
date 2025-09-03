package discovery

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

type RegisterOptions struct {
	ServiceName     string
	ServicePort     int
	ServiceID       string // 可选；为空则自动生成
	HealthCheckType string // "grpc" or "http"
}

func RegisterServiceToConsul(opts RegisterOptions) {
	consulAddr := os.Getenv("CONSUL_AGENT_ADDR")
	if consulAddr == "" {
		consulAddr = "127.0.0.1:8500"
	}

	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("❌ failed to create Consul client: %v", err)
	}

	serviceID := opts.ServiceID
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s", opts.ServiceName, uuid.New().String())
	}

	var check *api.AgentServiceCheck
	var tags []string

	switch opts.HealthCheckType {
	case "grpc":
		check = &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", opts.ServiceName, opts.ServicePort),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "1m",
		}
		tags = []string{"grpc"}
	case "http":
		check = &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", opts.ServiceName, opts.ServicePort),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "1m",
		}
		tags = []string{"http"}
	default:
		log.Fatalf("unknown health check type: %s. Must be 'grpc' or 'http'", opts.HealthCheckType)
	}

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    opts.ServiceName,
		Address: opts.ServiceName, // 与 docker-compose 中的 service 名相同
		Port:    opts.ServicePort,
		Tags:    tags,
		Check:   check,
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Fatalf("failed to register service with Consul: %v", err)
	}

	log.Printf("✅ Service %s registered to Consul with ID %s", opts.ServiceName, serviceID)
}

// GetServiceConnection creates a gRPC connection to a service via Consul
func GetServiceConnection(serviceName string) (*grpc.ClientConn, error) {
	consulAddr := os.Getenv("CONSUL_AGENT_ADDR")
	if consulAddr == "" {
		consulAddr = "127.0.0.1:8500"
	}

	log.Printf("[DISCOVERY] Attempting to connect to service: %s via Consul at %s", serviceName, consulAddr)

	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("[DISCOVERY] Failed to create Consul client: %v", err)
		return nil, fmt.Errorf("failed to create Consul client: %v", err)
	}

	// Query service from Consul
	log.Printf("[DISCOVERY] Querying Consul for service: %s", serviceName)
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("[DISCOVERY] Failed to query service %s from Consul: %v", serviceName, err)
		return nil, fmt.Errorf("failed to query service %s: %v", serviceName, err)
	}

	log.Printf("[DISCOVERY] Found %d healthy instances of service %s", len(services), serviceName)
	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	// Use the first healthy service instance
	service := services[0]
	address := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
	log.Printf("[DISCOVERY] Connecting to service %s at address: %s", serviceName, address)

	// Create gRPC connection
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("[DISCOVERY] Failed to establish gRPC connection to %s at %s: %v", serviceName, address, err)
		return nil, fmt.Errorf("failed to connect to service %s at %s: %v", serviceName, address, err)
	}

	log.Printf("[DISCOVERY] Successfully connected to service %s at %s", serviceName, address)
	return conn, nil
}
