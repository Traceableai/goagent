// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package empty

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FooClient is the client API for Foo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooClient interface {
	Bar(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error)
}

type fooClient struct {
	cc grpc.ClientConnInterface
}

func NewFooClient(cc grpc.ClientConnInterface) FooClient {
	return &fooClient{cc}
}

func (c *fooClient) Bar(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error) {
	out := new(wrapperspb.BoolValue)
	err := c.cc.Invoke(ctx, "/empty.Foo/Bar", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooServer is the server API for Foo service.
// All implementations must embed UnimplementedFooServer
// for forward compatibility
type FooServer interface {
	Bar(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error)
	mustEmbedUnimplementedFooServer()
}

// UnimplementedFooServer must be embedded to have forward compatible implementations.
type UnimplementedFooServer struct {
}

func (UnimplementedFooServer) Bar(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bar not implemented")
}
func (UnimplementedFooServer) mustEmbedUnimplementedFooServer() {}

// UnsafeFooServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooServer will
// result in compilation errors.
type UnsafeFooServer interface {
	mustEmbedUnimplementedFooServer()
}

func RegisterFooServer(s grpc.ServiceRegistrar, srv FooServer) {
	s.RegisterService(&Foo_ServiceDesc, srv)
}

func _Foo_Bar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooServer).Bar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/empty.Foo/Bar",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooServer).Bar(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Foo_ServiceDesc is the grpc.ServiceDesc for Foo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Foo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "empty.Foo",
	HandlerType: (*FooServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bar",
			Handler:    _Foo_Bar_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "empty.proto",
}
