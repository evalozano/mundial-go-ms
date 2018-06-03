package main

import (
	"fmt"

	"mundial-go-ms/registry"
	"mundial-go-ms/services/geo"
	"mundial-go-ms/tracing"
)

const geoSrvName = "srv-geo"

func runGeo(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("geo", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// service registry
	id, err := consul.Register(geoSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := geo.NewServer(tracer)
	return srv.Run(port)
}
