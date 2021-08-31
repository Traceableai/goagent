package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	pb "github.com/Traceableai/goagent/examples/internal/helloworld"
	"github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	cfg := config.Load()
	cfg.ServiceName = config.String("grpc-client")

	closer := goagent.Init(cfg)
	defer closer()

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			traceablegrpc.WrapUnaryClientInterceptor(
				otelgrpc.UnaryClientInterceptor(),
			),
		),
	)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %v", r.GetMessage())
}
