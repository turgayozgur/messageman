package main

import (
	"fmt"
	"net/url"

	"github.com/turgayozgur/messageman/internal/logging"
	"github.com/turgayozgur/messageman/internal/messaging"
	"github.com/turgayozgur/messageman/internal/waitfor"

	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/internal/metrics"

	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/messaging/rabbitmq"
	"github.com/turgayozgur/messageman/service"
)

var (
	workerRegistrars     = map[string]*messaging.WorkerRegistrar{}
	subscriberRegistrars = map[string]*messaging.SubscriberRegistrar{}
)

func main() {
	// init logger.
	logging.InitZerolog(config.Cfg.Logging.Level, config.Cfg.Logging.Humanize)

	// load configurations.
	if err := config.Load(); err != nil {
		log.Error().Msgf("failed to load config. %v", err.Error())
	}

	// create metric exporter.
	exporter := metrics.CreateExporter(config.Cfg.Metric.Enabled, config.Cfg.Metric.Exporter)

	// initialize messager to queue messages sent.
	m := rabbitmq.New(exporter)
	// wrapper
	w := &messaging.DefaultWrapper{}

	// check the sidecar mode and service count
	s := ""
	if config.IsSidecar() {
		log.Info().Msg("mode: sidecar")
		s = initSidecar(m)
	} else {
		log.Info().Msg("mode: gateway")
	}

	initConsumers(m, w)

	initRecover(m)

	service.NewServer(m, w, exporter, s).Listen()
}

func initConsumers(m messaging.Messager, w messaging.Wrapper) {
	go func() {
		if !config.IsSidecar() { // already waited on main method for sidecar mode.
			// wait for the connection to establish.
			waitfor.True(m.EnsureCanConnect)
		}
		for _, s := range config.Cfg.Events {
			// register subscribers if any.
			sr := messaging.NewSubscriberRegistrar(m, w, s)
			sr.RegisterSubscribers()
			subscriberRegistrars[s.Name] = sr
		}
		for _, s := range config.Cfg.Queues {
			// register workers if any.
			wr := messaging.NewWorkerRegistrar(m, w, s)
			wr.RegisterWorker()
			workerRegistrars[s.Name] = wr
		}
	}()
}

func initRecover(m messaging.Messager) {
	go func() {
		ch := make(chan string)
		for {
			name := <-m.NotifyRecover(ch)
			if v, ok := subscriberRegistrars[name]; ok {
				v.RegisterSubscribers()
			}
			if v, ok := workerRegistrars[name]; ok {
				v.RegisterWorker()
			}
		}
	}()
}

func initSidecar(m messaging.Messager) (service string) {
	var s config.ServiceConfig
	// Check the configured services.
	if len(config.Cfg.Queues) > 0 {
		s = config.Cfg.Queues[0].Worker
	} else if len(config.Cfg.Events) > 0 {
		s = config.Cfg.Events[0].Subscribers[0]
	}
	// If are on the sidecar mode, wait the service to be ready is the best idea to continue.
	readinessPath := s.Readiness.Path
	if readinessPath == "" {
		log.Error().Msg("readiness path is required for sidecar mode.")
		return
	}
	// wait for the main API is up and running.
	u, _ := url.Parse(s.Url)
	waitfor.API(fmt.Sprintf("%s%s:%s%s", u.Scheme, u.Host, u.Port(), readinessPath))
	// wait for the connection to establish.
	waitfor.True(m.EnsureCanConnect)
	return service
}
