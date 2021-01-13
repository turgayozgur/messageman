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

// Publish push message to subscribers.
func (s *Server) PublishREST(ctx *fasthttp.RequestCtx) {
	eventName := string(ctx.QueryArgs().Peek("name"))

	if eventName == "" {
		s.badRequest(ctx, "\"name\" parameter is required.")
		return
	}

	body := ctx.PostBody()
	if body == nil || len(body) == 0 {
		s.badRequest(ctx, "The request body is required.")
		return
	}

	var publisher string
	if s.mainAPI != "" {
		publisher = s.mainAPI
	} else {
		// We can get the publisher name from header if the x-service-name header provided.
		publisher = string(ctx.Request.Header.Peek("x-service-name"))
	}

	err := s.messager.Publish(publisher, eventName, body)
	if err != nil {
		s.error(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	s.write(ctx, fasthttp.StatusOK, nil)
}

// Publish push message to subscribers by using gRPC.
func (s *Server) Publish(ctx context.Context, in *pb.PublishRequest) (*empty.Empty, error) {
	eventName := in.Name

	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "The \"name\" field is required.")
	}

	body := in.Message
	if len(body) == 0 {
		return nil, status.Error(codes.InvalidArgument, "The \"message\" field is required.")
	}

	var publisher string
	if s.mainAPI != "" {
		publisher = s.mainAPI
	} else if md, ok := metadata.FromIncomingContext(ctx); ok {
		// We can get the service name from header if the x-service-name header provided.
		h := md.Get("x-service-name")
		if len(h) > 0 {
			publisher = h[0]
		}
	}

	err := s.messager.Publish(publisher, eventName, body)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &empty.Empty{}, nil
}
