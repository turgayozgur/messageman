package messaging

import (
	"bytes"
	"context"
	"fmt"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/waitfor"
)

const ContentType = "application/json"

var gRPCClients = map[string]*grpc.ClientConn{}

// WorkerRegistrar .
type WorkerRegistrar struct {
	messager   Messager
	httpClient *http.Client
}

func NewWorkerRegistrar(messager Messager) *WorkerRegistrar {
	return &WorkerRegistrar{
		messager: messager,
		// Clients and Transports are safe for concurrent use by multiple goroutines
		// and for efficiency should only be created once and re-used.
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// RegisterWorkers to register given endpoints to queues via environment variables.
func (wr *WorkerRegistrar) RegisterWorkers(service config.ServiceConfig) {
	if service.Workers == nil {
		return
	}

	if !config.IsSidecar() { // already waited on main method for sidecar mode.
		// wait for the connection to establish.
		waitfor.True(wr.messager.EnsureCanConnect, service.Name)
	}

	for _, worker := range service.Workers {
		wr.registerWorker(service, worker)
	}
}

func (wr *WorkerRegistrar) registerWorker(service config.ServiceConfig, c config.WorkerConfig) {
	path := c.Path
	queue := c.Queue
	mainAPI := service.Name
	url := fmt.Sprintf("%s%s", service.Url, path)

	if c.Type == "gRPC" {
		if err := connGRPC(mainAPI, service.Url); err != nil {
			log.Error().Msgf("Failed to connect to gRPC endpoint %s for the service %s", service.Url, mainAPI)
			return
		}
	}

	err := wr.messager.Receive(service.Name, queue, func(body []byte) bool {
		log.Debug().Str("body", string(body)).Msg("Job received")
		var ok bool
		if c.Type == "gRPC" {
			ok = wr.receiveGRPC(mainAPI, queue, body)
		} else {
			ok = wr.receiveREST(mainAPI, url, queue, body)
		}
		if !ok {
			return ok
		}
		log.Debug().Str("body", string(body)).Msg("Job succeeded")
		return true
	})
	if err != nil {
		log.Fatal().Err(err).Msg("")
	} else {
		log.Info().Str("queue", queue).Str("type", c.Type).Msgf("Worker '%s' registered", url)
	}
}

func (wr *WorkerRegistrar) receiveREST(mainAPI string, url string, queue string, body []byte) bool {
	response, err := wr.httpClient.Post(url, ContentType, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Str("body", string(body)).Str("mainApi", mainAPI).Str("queue", queue).
			Msgf("Job failed. An error occurred on http post. url:%s", url)
		return false
	}
	if response.StatusCode >= 300 {
		log.Error().Str("body", string(body)).Str("mainApi", mainAPI).Str("queue", queue).
			Msgf("Job failed. Non success status code %d on http post to worker. url:%s", response.StatusCode, url)
		return false
	}
	return true
}

func (wr *WorkerRegistrar) receiveGRPC(mainAPI string, queue string, body []byte) bool {
	c := pb.NewWorkerServiceClient(gRPCClients[mainAPI])
	_, err := c.Receive(context.Background(), &pb.ReceiveRequest{
		Name:    queue,
		Message: body,
	})
	if err != nil {
		l := log.Error().Str("body", string(body)).Str("mainApi", mainAPI).Str("queue", queue)
		if s, ok := status.FromError(err); ok {
			l.Msgf("Job failed. Non success gRPC status code %d on http post to worker. message:%s", s.Code(), s.Message())
		}
		l.Msgf("Job failed. Unknown error from gRPC endpoint. %v", err)
		return false
	}
	return true
}

func connGRPC(name string, addr string) error {
	ctx := context.Background()
	var err error

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	cnn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return err
	}

	gRPCClients[name] = cnn

	return nil
}
