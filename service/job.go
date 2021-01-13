package service

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Queue push message to workers.
func (s *Server) QueueREST(ctx *fasthttp.RequestCtx) {
	queueName := string(ctx.QueryArgs().Peek("name"))

	if queueName == "" {
		s.badRequest(ctx, "\"name\" parameter is required.")
		return
	}

	body := ctx.PostBody()
	if body == nil || len(body) == 0 {
		s.badRequest(ctx, "the request body is required.")
		return
	}

	var service string
	if s.mainAPI != "" {
		service = s.mainAPI
	} else {
		// We can get the main api name from header if the x-service-name header provided.
		service = string(ctx.Request.Header.Peek("x-service-name"))
	}

	err := s.messager.Queue(service, queueName, body)
	if err != nil {
		s.error(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	s.write(ctx, fasthttp.StatusOK, nil)
}

// Queue push message to workers by using gRPC.
func (s *Server) Queue(ctx context.Context, in *pb.QueueRequest) (*empty.Empty, error) {
	queueName := in.Name

	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "the \"name\" field is required.")
	}

	body := in.Message
	if len(body) == 0 {
		return nil, status.Error(codes.InvalidArgument, "the \"message\" field is required.")
	}

	var service string
	if s.mainAPI != "" {
		service = s.mainAPI
	} else if md, ok := metadata.FromIncomingContext(ctx); ok {
		// We can get the service name from header if the x-service-name header provided.
		h := md.Get("x-service-name")
		if len(h) > 0 {
			service = h[0]
		}
	}

	err := s.messager.Queue(service, queueName, body)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &empty.Empty{}, nil
}
