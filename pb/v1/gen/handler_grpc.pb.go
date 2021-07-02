// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package messageman

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// HandlerServiceClient is the client API for HandlerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HandlerServiceClient interface {
	Handle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type handlerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHandlerServiceClient(cc grpc.ClientConnInterface) HandlerServiceClient {
	return &handlerServiceClient{cc}
}

func (c *handlerServiceClient) Handle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/messageman.v1.HandlerService/Handle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HandlerServiceServer is the server API for HandlerService service.
// All implementations must embed UnimplementedHandlerServiceServer
// for forward compatibility
type HandlerServiceServer interface {
	Handle(context.Context, *HandleRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedHandlerServiceServer()
}

// UnimplementedHandlerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHandlerServiceServer struct {
}

func (UnimplementedHandlerServiceServer) Handle(context.Context, *HandleRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Handle not implemented")
}
func (UnimplementedHandlerServiceServer) mustEmbedUnimplementedHandlerServiceServer() {}

// UnsafeHandlerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HandlerServiceServer will
// result in compilation errors.
type UnsafeHandlerServiceServer interface {
	mustEmbedUnimplementedHandlerServiceServer()
}

func RegisterHandlerServiceServer(s grpc.ServiceRegistrar, srv HandlerServiceServer) {
	s.RegisterService(&HandlerService_ServiceDesc, srv)
}

func _HandlerService_Handle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerServiceServer).Handle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messageman.v1.HandlerService/Handle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerServiceServer).Handle(ctx, req.(*HandleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HandlerService_ServiceDesc is the grpc.ServiceDesc for HandlerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HandlerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "messageman.v1.HandlerService",
	HandlerType: (*HandlerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Handle",
			Handler:    _HandlerService_Handle_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/v1/handler.proto",
}
