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
		s.badRequest(ctx, "The request body is required.")
		return
	}

	var mainAPI string
	if s.mainAPI != "" {
		mainAPI = s.mainAPI
	} else {
		// We can get the main api name from header if the x-sender-name header provided.
		mainAPI = string(ctx.Request.Header.Peek("x-sender-name"))
	}

	err := s.messager.Send(mainAPI, queueName, body)
	if err != nil {
		s.error(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	s.write(ctx, fasthttp.StatusOK, &ResponseModel{Message: "Successfully queued."})
}

// Queue push message to workers by using gRPC.
func (s *Server) Queue(ctx context.Context, in *pb.QueueRequest) (*empty.Empty, error) {
	queueName := in.Name

	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "The \"name\" field is required.")
	}

	body := in.Message
	if len(body) == 0 {
		return nil, status.Error(codes.InvalidArgument, "The \"message\" field is required.")
	}

	var mainAPI string
	if s.mainAPI != "" {
		mainAPI = s.mainAPI
	} else if md, ok := metadata.FromIncomingContext(ctx); ok {
		// We can get the main api name from header if the x-sender-name header provided.
		h := md.Get("x-sender-name")
		if len(h) > 0 {
			mainAPI = h[0]
		}
	}

	err := s.messager.Send(mainAPI, queueName, body)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, status.Errorf(codes.Unimplemented, "method Queue not implemented")
}
