package messaging

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/waitfor"
	"github.com/rs/zerolog/log"
)

const ContentType = "application/json"

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

func (wr *WorkerRegistrar) registerWorker(service config.ServiceConfig, worker config.WorkerConfig) {
	path := worker.Path
	queue := worker.Queue
	mainAPI := service.Name
	url := fmt.Sprintf("%s%s", service.Url, path)

	err := wr.messager.Receive(service.Name, queue, func(body []byte) bool {
		log.Debug().Str("body", string(body)).Msg("Job received")
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
		log.Debug().Str("body", string(body)).Msg("Job done")
		return true
	})
	if err != nil {
		log.Fatal().Err(err).Msg("")
	} else {
		log.Info().Str("queue", queue).Msgf("Worker '%s' registered", url)
	}
}
