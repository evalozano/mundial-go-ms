package search

import (
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	geo "mundial-go-ms/services/geo/proto"
	rate "mundial-go-ms/services/rate/proto"
	pb "mundial-go-ms/services/search/proto"
	opentracing "github.com/opentracing/opentracing-go"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServer returns a new server
func NewServer(gc geo.GeoClient, rc rate.RateClient, tr opentracing.Tracer) *Server {
	return &Server{
		geoClient:  gc,
		rateClient: rc,
		tracer:     tr,
	}
}

// Server implements the search service
type Server struct {
	geoClient  geo.GeoClient
	rateClient rate.RateClient
	tracer     opentracing.Tracer
}

// Run starts the server
func (s *Server) Run(port int) error {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.tracer),
		),
	)
	pb.RegisterSearchServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return srv.Serve(lis)
}

// Nearby returns ids of nearby pubs ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// find nearby pubs
	nearby, err := s.geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		log.Fatalf("nearby error: %v", err)
	}

	// find rates for pubs
	rates, err := s.rateClient.GetRates(ctx, &rate.Request{
		PubIds: nearby.PubIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		log.Fatalf("rates error: %v", err)
	}

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.PubIds = append(res.PubIds, ratePlan.PubId)
	}
	return res, nil
}
