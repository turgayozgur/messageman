package service

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/turgayozgur/messageman/config"
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

	var err error
	if body, err = s.wrapBodyREST(ctx, body); err != nil {
		s.error(ctx, fasthttp.StatusInternalServerError, "failed to wrap message.")
	}

	if err = s.messager.Queue(service, queueName, body); err != nil {
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

	md, mdOk := metadata.FromIncomingContext(ctx)

	var service string
	if s.mainAPI != "" {
		service = s.mainAPI
	} else if mdOk {
		// We can get the service name from header if the x-service-name header provided.
		h := md.Get("x-service-name")
		if len(h) > 0 {
			service = h[0]
		}
	}

	var err error
	if mdOk {
		if body, err = s.wrapBodyGRPC(mdOk, md, body); err != nil {
			return nil, status.Error(codes.Internal, "failed to wrap message.")
		}
	}

	if err = s.messager.Queue(service, queueName, body); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &empty.Empty{}, nil
}

func (s *Server) wrapBodyREST(ctx *fasthttp.RequestCtx, body []byte) ([]byte, error) {
	if config.Cfg.Proxy != nil && config.Cfg.Proxy.Headers != nil {
		headers := make(map[string][]byte, len(config.Cfg.Proxy.Headers))
		for _, v := range config.Cfg.Proxy.Headers {
			headers[v] = ctx.Request.Header.Peek(v)
		}
		return s.wrapper.Wrap(body, headers)
	}
	return s.wrapper.Wrap(body, map[string][]byte{})
}

func (s *Server) wrapBodyGRPC(mdOk bool, md metadata.MD, body []byte) ([]byte, error) {
	if mdOk && config.Cfg.Proxy != nil && config.Cfg.Proxy.Headers != nil {
		headers := make(map[string][]byte, len(config.Cfg.Proxy.Headers))
		for _, v := range config.Cfg.Proxy.Headers {
			h := md.Get(v)
			if len(h) > 0 {
				headers[v] = []byte(h[0])
			}
		}
		return s.wrapper.Wrap(body, headers)
	}
	return s.wrapper.Wrap(body, map[string][]byte{})
}
