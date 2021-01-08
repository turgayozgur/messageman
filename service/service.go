package service

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/messaging"
	"github.com/turgayozgur/messageman/internal/metrics"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"net"
)

// Server contains all that is needed to respond to incoming requests, like a database. Other services like a mail
type Server struct {
	pb.UnimplementedJobServiceServer
	pb.UnimplementedPublishServiceServer
	messager messaging.Messager
	exporter metrics.Exporter
	mainAPI  string
}

// NewServer initializes the service with the given Database, and sets up appropriate routes.
func NewServer(messager messaging.Messager, exporter metrics.Exporter, mainAPI string) *Server {
	server := &Server{
		messager: messager,
		exporter: exporter,
		mainAPI:  mainAPI,
	}
	return server
}

func (s *Server) Listen() {
	// listen gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Cfg.GRPCPort))
	if err != nil {
		log.Fatal().Err(err)
	}
	gSrv := grpc.NewServer()
	pb.RegisterJobServiceServer(gSrv, s)
	pb.RegisterPublishServiceServer(gSrv, s)
	go func() {
		log.Info().Msgf("Now, gRPC listening on: http://localhost:%s", config.Cfg.GRPCPort)
		if err := gSrv.Serve(lis); err != nil {
			log.Fatal().Err(err)
		}
	}()

	// listen REST
	m := func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if rc := recover(); rc != nil {
				message := fmt.Sprintf("Panic during request. %+v", rc)
				s.error(ctx, fasthttp.StatusInternalServerError, message)
				log.Error().Msg(message)
			}
		}()
		switch string(ctx.Path()) {
		case "/":
		case "/healthz":
			s.healthz(ctx)
		case "/queue":
			s.QueueREST(ctx)
		case "/publish":
			s.PublishREST(ctx)
		case "/metrics":
			s.exporter.Handle(ctx)
		default:
			s.notFound(ctx)
		}
	}

	log.Info().Msgf("Now, listening on: http://localhost:%s", config.Cfg.Port)
	if err := fasthttp.ListenAndServe(":"+config.Cfg.Port, m); err != nil {
		log.Fatal().Err(err)
	}
}

// healthz handles HTTP requests to know the messageman is healthy or not.
func (s *Server) healthz(ctx *fasthttp.RequestCtx) {
	s.write(
		ctx,
		fasthttp.StatusOK,
		&ResponseModel{
			Message: "Welcome to messageman! The ultimate message manager proxy.",
		},
	)
}

// Helpers
func (s *Server) write(ctx *fasthttp.RequestCtx, statusCode int, response interface{}) {
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetStatusCode(statusCode)

	if response == nil {
		return
	}

	if err := json.NewEncoder(ctx).Encode(response); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}

func (s *Server) error(ctx *fasthttp.RequestCtx, statusCode int, message string) {
	s.write(
		ctx,
		statusCode,
		&ResponseModel{
			Message: message,
		},
	)
}

func (s *Server) badRequest(ctx *fasthttp.RequestCtx, message string) {
	s.error(ctx, fasthttp.StatusBadRequest, message)
}

func (s *Server) notFound(ctx *fasthttp.RequestCtx) {
	s.error(ctx, fasthttp.StatusNotFound, "404")
}

// ResponseModel is returned by our service when an error occurs.
type ResponseModel struct {
	Message string `json:"message"`
}
