// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package messageman

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// JobServiceClient is the client API for JobService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JobServiceClient interface {
	Queue(ctx context.Context, in *QueueRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type jobServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewJobServiceClient(cc grpc.ClientConnInterface) JobServiceClient {
	return &jobServiceClient{cc}
}

func (c *jobServiceClient) Queue(ctx context.Context, in *QueueRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/messageman.v1.JobService/Queue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JobServiceServer is the server API for JobService service.
// All implementations must embed UnimplementedJobServiceServer
// for forward compatibility
type JobServiceServer interface {
	Queue(context.Context, *QueueRequest) (*empty.Empty, error)
	mustEmbedUnimplementedJobServiceServer()
}

// UnimplementedJobServiceServer must be embedded to have forward compatible implementations.
type UnimplementedJobServiceServer struct {
}

func (UnimplementedJobServiceServer) Queue(context.Context, *QueueRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Queue not implemented")
}
func (UnimplementedJobServiceServer) mustEmbedUnimplementedJobServiceServer() {}

// UnsafeJobServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JobServiceServer will
// result in compilation errors.
type UnsafeJobServiceServer interface {
	mustEmbedUnimplementedJobServiceServer()
}

func RegisterJobServiceServer(s grpc.ServiceRegistrar, srv JobServiceServer) {
	s.RegisterService(&_JobService_serviceDesc, srv)
}

func _JobService_Queue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).Queue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messageman.v1.JobService/Queue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).Queue(ctx, req.(*QueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _JobService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "messageman.v1.JobService",
	HandlerType: (*JobServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Queue",
			Handler:    _JobService_Queue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/job.proto",
}
