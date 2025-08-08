package discovery

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
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
			GRPC:                           fmt.Sprintf("%s:%d", opts.ServiceName, opts.ServicePort),
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
