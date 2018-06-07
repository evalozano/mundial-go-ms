package profile

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"mundial-go-ms/data"
	pb "mundial-go-ms/services/profile/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServer returns a new server
func NewServer(tr opentracing.Tracer) *Server {
	return &Server{
		tracer: tr,
		pubs: loadProfiles("data/pubs.json"),
	}
}

// Server implements the profile service
type Server struct {
	pubs map[string]*pb.Pub
	tracer opentracing.Tracer
}

// Run starts the server
func (s *Server) Run(port int) error {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.tracer),
		),
	)
	pb.RegisterProfileServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return srv.Serve(lis)
}

// GetProfiles returns pub profiles for requested IDs
func (s *Server) GetProfiles(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	for _, i := range req.PubIds {
		res.Pubs = append(res.Pubs, s.pubs[i])
	}
	return res, nil
}

// loadProfiles loads pub profiles from a JSON file.
func loadProfiles(path string) map[string]*pb.Pub {
	file := data.MustAsset(path)

	// unmarshal json profiles
	pubs := []*pb.Pub{}
	if err := json.Unmarshal(file, &pubs); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*pb.Pub)
	for _, pub := range pubs {
		profiles[pub.Id] = pub
	}

	return profiles
}
