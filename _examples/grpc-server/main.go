package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Traceableai/goagent"
	pb "github.com/Traceableai/goagent/_examples/internal/helloworld"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

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
	cfg := config.LoadFromFile("./config.yaml")

	closer := goagent.Init(cfg)
	defer closer()

	// gosec: "G102: Binds to all network interfaces". Ignore this error.
	lis, err := net.Listen("tcp", port) // #nosec
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			traceablegrpc.UnaryServerInterceptor(),
		),
	)

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
