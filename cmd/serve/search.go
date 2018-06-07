package main

import (
	"fmt"

	"mundial-go-ms/dialer"
	"mundial-go-ms/registry"
	geo "mundial-go-ms/services/geo/proto"
	rate "mundial-go-ms/services/rate/proto"
	"mundial-go-ms/services/search"
	"mundial-go-ms/tracing"
)

const searchSrvName = "srv-search"

func runSearch(port int, consul *registry.Client, jaegeraddr string) error {
	// We initialize the trace service 
	tracer, err := tracing.Init("search", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// dial geo srv
	gc, err := dialer.Dial(
		geoSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	// dial rate srv
	rc, err := dialer.Dial(
		rateSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	// service registry
	id, err := consul.Register(searchSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := search.NewServer(
		geo.NewGeoClient(gc),
		rate.NewRateClient(rc),
		tracer,
	)
	return srv.Run(port)
}
