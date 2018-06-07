package main

import (
	"fmt"

	"mundial-go-ms/registry"
	"mundial-go-ms/services/profile"
	"mundial-go-ms/tracing"
)

const profileSrvName = "srv-profile"

//
func runProfile(port int, consul *registry.Client, jaegeraddr string) error {
	// We initialize the trace service
	tracer, err := tracing.Init("profile", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// service registry
	id, err := consul.Register(profileSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}

	// deregister of consult
	defer consul.Deregister(id)

	srv := profile.NewServer(tracer)
	return srv.Run(port)
}
