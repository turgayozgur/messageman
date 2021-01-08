package messaging

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/waitfor"
	pb "github.com/turgayozgur/messageman/pb/v1/gen"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

// SubscriberRegistrar .
type SubscriberRegistrar struct {
	messager   Messager
	httpClient *http.Client
}

func NewSubscriberRegistrar(messager Messager) *SubscriberRegistrar {
	return &SubscriberRegistrar{
		messager: messager,
		// Clients and Transports are safe for concurrent use by multiple goroutines
		// and for efficiency should only be created once and re-used.
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// RegisterSubscribers to register given endpoints to queues via environment variables.
func (s *SubscriberRegistrar) RegisterSubscribers(service config.ServiceConfig) {
	if service.Subscribers == nil {
		return
	}

	if !config.IsSidecar() { // already waited on main method for sidecar mode.
		// wait for the connection to establish.
		waitfor.True(s.messager.EnsureCanConnect, service.Name)
	}

	for _, subscriber := range service.Subscribers {
		s.registerSubscriber(service, subscriber)
	}
}

func (s *SubscriberRegistrar) registerSubscriber(service config.ServiceConfig, c config.SubscriberConfig) {
	path := c.Path
	eventName := c.Event
	subscriber := service.Name
	url := fmt.Sprintf("%s%s", service.Url, path)

	if c.Type == "gRPC" {
		if err := connGRPC(subscriber, service.Url); err != nil {
			log.Error().Msgf("Failed to connect to gRPC endpoint %s for the service %s", service.Url, subscriber)
			return
		}
	}

	err := s.messager.Subscribe(service.Name, eventName, func(body []byte) bool {
		log.Debug().Str("body", string(body)).Msg("Message received")
		var ok bool
		if c.Type == "gRPC" {
			ok = s.handleGRPC(subscriber, eventName, body)
		} else {
			ok = s.handleREST(subscriber, url, eventName, body)
		}
		if !ok {
			return ok
		}
		log.Debug().Str("body", string(body)).Msg("Successfully handled")
		return true
	})
	if err != nil {
		log.Fatal().Err(err).Msg("")
	} else {
		log.Info().Str("event", eventName).Str("type", c.Type).Msgf("Subscriber '%s' registered", url)
	}
}

func (s *SubscriberRegistrar) handleREST(subscriber string, url string, queue string, body []byte) bool {
	response, err := s.httpClient.Post(url, ContentType, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Str("body", string(body)).Str("subscriber", subscriber).Str("queue", queue).
			Msgf("Handle failed. An error occurred on http post. url:%s", url)
		return false
	}
	if response.StatusCode >= 300 {
		log.Error().Str("body", string(body)).Str("subscriber", subscriber).Str("queue", queue).
			Msgf("Handle failed. Non success status code %d on http post to subscriber. url:%s", response.StatusCode, url)
		return false
	}
	return true
}

func (s *SubscriberRegistrar) handleGRPC(subscriber string, eventName string, body []byte) bool {
	c := pb.NewHandlerServiceClient(gRPCClients[subscriber])
	_, err := c.Handle(context.Background(), &pb.HandleRequest{
		Name:    eventName,
		Message: body,
	})
	if err != nil {
		l := log.Error().Str("body", string(body)).Str("subscriber", subscriber).Str("event", eventName)
		if s, ok := status.FromError(err); ok {
			l.Msgf("Handle failed. Non success gRPC status code %d on http post to subscriber. message:%s", s.Code(), s.Message())
			return false
		}
		l.Msgf("Handle failed. Unknown error from gRPC endpoint. %v", err)
		return false
	}
	return true
}
