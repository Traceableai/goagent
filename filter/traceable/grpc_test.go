//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	pb "github.com/Traceableai/goagent/filter/traceable/internal/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestIsGRPC(t *testing.T) {
	assert.False(t, isGRPC(map[string][]string{}))
	assert.False(t, isGRPC(map[string][]string{"content-type": []string{}}))
	assert.False(t, isGRPC(map[string][]string{"content-type": []string{"text/html"}}))
	assert.False(t, isGRPC(map[string][]string{"content-type": []string{"application/json"}}))
	assert.True(t, isGRPC(map[string][]string{"content-type": []string{"application/grpc+proto"}}))
}

type FooServer struct {
	pb.UnimplementedFooServer
}

func (*FooServer) Bar(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata")
	}

	return &wrapperspb.BoolValue{Value: isGRPC(md)}, nil
}

func createDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	pb.RegisterFooServer(server, &FooServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestIsGRPCSuccess(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithContextDialer(createDialer()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewFooClient(conn)
	req := &emptypb.Empty{}

	res, err := client.Bar(ctx, req)
	require.NoError(t, err)
	require.True(t, res.Value)
}
