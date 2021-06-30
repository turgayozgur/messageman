package messaging

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/config"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

// SubscriberRegistrar .
type SubscriberRegistrar struct {
	messager   Messager
	wrapper    Wrapper
	cfg        *config.EventConfig
	httpClient *http.Client
}

func NewSubscriberRegistrar(m Messager, w Wrapper, cfg *config.EventConfig) *SubscriberRegistrar {
	return &SubscriberRegistrar{
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

// RegisterSubscribers to register given endpoints to queues via environment variables.
func (s *SubscriberRegistrar) RegisterSubscribers() {
	if s.cfg.Subscribers == nil {
		return
	}

	for _, c := range s.cfg.Subscribers {
		s.registerSubscriber(c)
	}
}

func (s *SubscriberRegistrar) registerSubscriber(c config.ServiceConfig) {
	name := s.cfg.Name
	service := c.Name
	url := c.Url

	if c.Type == "gRPC" {
		if err := connGRPC(service, url); err != nil {
			log.Error().Msgf("failed to connect to gRPC endpoint %s for the service %s", url, service)
			return
		}
	}

	err := s.messager.Subscribe(service, name, func(body []byte) bool {
		log.Debug().Str("body", string(body)).Msg("message received")
		var ok bool
		if c.Type == "gRPC" {
			ok = s.handleGRPC(service, name, body)
		} else {
			ok = s.handleREST(service, url, name, body)
		}
		if !ok {
			return ok
		}
		log.Debug().Str("body", string(body)).Msg("successfully handled")
		return true
	})
	if err != nil {
		log.Error().Err(err).Msg("")
	} else {
		log.Info().Str("name", name).Str("service", service).Str("type", c.Type).Msg("subscriber registered")
	}
}

func (s *SubscriberRegistrar) handleREST(service string, url string, name string, body []byte) bool {
	response, err := doRest(s.wrapper, s.httpClient, url, body)
	if err != nil {
		log.Error().Err(err).Str("body", string(body)).Str("v", service).Str("name", name).
			Msgf("handle failed. An error occurred on http post. url:%s", url)
		return false
	}
	if response.StatusCode >= 300 {
		log.Error().Str("body", string(body)).Str("service", service).Str("name", name).
			Msgf("handle failed. Non success status code %d on http post to subscriber. url:%s", response.StatusCode, url)
		return false
	}
	return true
}

func (s *SubscriberRegistrar) handleGRPC(service, name string, body []byte) bool {
	err := doGRPC(s.wrapper, body, func(ctx context.Context) error {
		c := pb.NewEventServiceClient(gRPCClients[service])
		_, err := c.Handle(ctx, &pb.HandleRequest{
			Name:    name,
			Message: body,
		})
		return err
	})
	if err != nil {
		l := log.Error().Str("body", string(body)).Str("service", service).Str("name", name)
		if s, ok := status.FromError(err); ok {
			l.Msgf("handle failed. Non success gRPC status code %d on http post to subscriber. message:%s", s.Code(), s.Message())
			return false
		}
		l.Msgf("handle failed. Unknown error from gRPC endpoint. %v", err)
		return false
	}
	return true
}
