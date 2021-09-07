package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	pb "github.com/Traceableai/goagent/examples/internal/helloworld"
	"github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %q", in.Name)
	return &pb.HelloReply{Message: fmt.Sprintf("hello %s", in.Name)}, nil
}

func main() {
	cfg := config.Load()
	cfg.Tracing.ServiceName = config.String("grpc-server")

	closer := goagent.Init(cfg)
	defer closer()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			traceablegrpc.WrapUnaryServerInterceptor(
				otelgrpc.UnaryServerInterceptor(),
				&traceablegrpc.Options{},
			),
		),
	)

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
