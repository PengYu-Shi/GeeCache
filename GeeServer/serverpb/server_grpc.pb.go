// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.0--rc1
// source: server.proto

package __

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GeeServer_Get_FullMethodName = "/serverpb.GeeServer/Get"
)

// GeeServerClient is the client API for GeeServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GeeServerClient interface {
	Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type geeServerClient struct {
	cc grpc.ClientConnInterface
}

func NewGeeServerClient(cc grpc.ClientConnInterface) GeeServerClient {
	return &geeServerClient{cc}
}

func (c *geeServerClient) Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, GeeServer_Get_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GeeServerServer is the server API for GeeServer service.
// All implementations must embed UnimplementedGeeServerServer
// for forward compatibility.
type GeeServerServer interface {
	Get(context.Context, *Request) (*Response, error)
	mustEmbedUnimplementedGeeServerServer()
}

// UnimplementedGeeServerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGeeServerServer struct{}

func (UnimplementedGeeServerServer) Get(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedGeeServerServer) mustEmbedUnimplementedGeeServerServer() {}
func (UnimplementedGeeServerServer) testEmbeddedByValue()                   {}

// UnsafeGeeServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GeeServerServer will
// result in compilation errors.
type UnsafeGeeServerServer interface {
	mustEmbedUnimplementedGeeServerServer()
}

func RegisterGeeServerServer(s grpc.ServiceRegistrar, srv GeeServerServer) {
	// If the following call pancis, it indicates UnimplementedGeeServerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GeeServer_ServiceDesc, srv)
}

func _GeeServer_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GeeServerServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GeeServer_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GeeServerServer).Get(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

// GeeServer_ServiceDesc is the grpc.ServiceDesc for GeeServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GeeServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "serverpb.GeeServer",
	HandlerType: (*GeeServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _GeeServer_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}
