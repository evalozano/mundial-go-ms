package main

import (
	"fmt"

	"mundial-go-ms/dialer"
	"mundial-go-ms/registry"
	"mundial-go-ms/services/frontend"
	profile "mundial-go-ms/services/profile/proto"
	search "mundial-go-ms/services/search/proto"
	"mundial-go-ms/tracing"
)

func runFrontend(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("frontend", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// dial search srv
	sc, err := dialer.Dial(
		searchSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	// dial profile srv
	pc, err := dialer.Dial(
		profileSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	srv := frontend.NewServer(
		search.NewSearchClient(sc),
		profile.NewProfileClient(pc),
		tracer,
	)
	return srv.Run(port)
}
