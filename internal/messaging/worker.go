package messaging

import (
	"context"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/config"
)

const ContentType = "application/json"

var gRPCClients = map[string]*grpc.ClientConn{}

// WorkerRegistrar .
type WorkerRegistrar struct {
	messager   Messager
	wrapper    Wrapper
	cfg        *config.QueueConfig
	httpClient *http.Client
}

func NewWorkerRegistrar(m Messager, w Wrapper, cfg *config.QueueConfig) *WorkerRegistrar {
	return &WorkerRegistrar{
		messager: m,
		wrapper:  w,
		cfg:      cfg,
		// Clients and Transports are safe for concurrent use by multiple goroutines
		// and for efficiency should only be created once and re-used.
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// RegisterWorkers to register given endpoints to queues via environment variables.
func (wr *WorkerRegistrar) RegisterWorker() {
	wr.registerWorker()
}

func (wr *WorkerRegistrar) registerWorker() {
	cfg := wr.cfg
	url := cfg.Worker.Url
	name := cfg.Name
	service := cfg.Worker.Name

	if cfg.Worker.Type == "gRPC" {
		if err := connGRPC(service, url); err != nil {
			log.Error().Msgf("failed to connect to gRPC endpoint %s for the service %s", url, service)
			return
		}
	}

	err := wr.messager.Work(service, name, func(body []byte) bool {
		log.Debug().Str("body", string(body)).Msg("job received")
		var ok bool
		if cfg.Worker.Type == "gRPC" {
			ok = wr.receiveGRPC(service, name, body)
		} else {
			ok = wr.receiveREST(service, url, name, body)
		}
		if !ok {
			return ok
		}
		log.Debug().Str("body", string(body)).Msg("job succeeded")
		return true
	})
	if err != nil {
		log.Error().Err(err).Msg("")
	} else {
		log.Info().Str("name", name).Str("service", service).Str("type", cfg.Worker.Type).Msg("worker registered")
	}
}

func (wr *WorkerRegistrar) receiveREST(service string, url string, name string, body []byte) bool {
	response, err := doRest(wr.wrapper, wr.httpClient, url, body)
	if err != nil {
		log.Error().Err(err).Str("body", string(body)).Str("service", service).Str("name", name).
			Msgf("job failed. An error occurred on http post. url:%s", url)
		return false
	}
	if response.StatusCode >= 300 {
		log.Error().Str("body", string(body)).Str("service", service).Str("name", name).
			Msgf("job failed. Non success status code %d on http post to worker. url:%s", response.StatusCode, url)
		return false
	}
	return true
}

func (wr *WorkerRegistrar) receiveGRPC(service string, name string, body []byte) bool {
	err := doGRPC(wr.wrapper, body, func(ctx context.Context, b []byte) error {
		c := pb.NewWorkerServiceClient(gRPCClients[service])
		_, err := c.Receive(ctx, &pb.ReceiveRequest{
			Name:    name,
			Message: b,
		})
		return err
	})
	if err != nil {
		l := log.Err(err).Str("body", string(body)).Str("service", service).Str("name", name)
		if s, ok := status.FromError(err); ok {
			l.Msgf("job failed. Non success gRPC status code %d on http post to worker. message:%s", s.Code(), s.Message())
		}
		l.Msgf("job failed. Unknown error from gRPC endpoint. %v", err)
		return false
	}
	return true
}

func connGRPC(service string, addr string) error {
	if _, ok := gRPCClients[service]; ok {
		return nil
	}

	ctx := context.Background()
	var err error

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	cnn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return err
	}

	gRPCClients[service] = cnn

	return nil
}
