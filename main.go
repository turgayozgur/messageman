package main

import (
	"fmt"

	"github.com/turgayozgur/messageman/internal/logging"
	"github.com/turgayozgur/messageman/internal/messaging"
	"github.com/turgayozgur/messageman/internal/waitfor"

	"github.com/rs/zerolog/log"
	"github.com/turgayozgur/messageman/internal/metrics"

	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/messaging/rabbitmq"
	"github.com/turgayozgur/messageman/service"
)

func main() {
	// init logger.
	logging.InitZerolog(config.Cfg.Logging.Level, config.Cfg.Logging.Humanize)

	// load configurations.
	if err := config.Load(); err != nil {
		log.Error().Msgf("Failed to load config. %v", err.Error())
	}

	// check have we any registered service
	if config.Cfg.Services == nil || len(config.Cfg.Services) == 0 {
		log.Warn().Msg("No any registered service found. Please check your configuration file")
	}

	// create metric exporter.
	exporter := metrics.CreateExporter(config.Cfg.Metric.Enabled, config.Cfg.Metric.Exporter)

	// initialize messager to queue messages sent.
	messager := rabbitmq.New(exporter)

	// check the sidecar mode and service count
	mainAPI := ""
	if config.IsSidecar() {
		log.Info().Msg("Mode: sidecar")
		mainAPI = initSidecar(messager)
	} else {
		log.Info().Msg("Mode: gateway")
	}

	// initialize services
	for _, s := range config.Cfg.Services {
		// register workers if any.
		wr := messaging.NewWorkerRegistrar(messager)
		go wr.RegisterWorkers(s)
	}

	service.NewServer(messager, exporter, mainAPI).Listen()
}

func initSidecar(messager messaging.Messager) (mainAPI string) {
	// Check the configured services.
	if config.Cfg.Services == nil || len(config.Cfg.Services) > 1 {
		log.Error().Msg("You need to register a service for the mode 'sidecar'. Only one service allowed for the mode 'sidecar'")
		return
	}
	// If are on the sidecar mode, wait the service to be ready is the best idea to continue.
	readinessPath := config.Cfg.Services[0].Readiness.Path
	if readinessPath == "" {
		log.Error().Msg("Readiness path is required for sidecar mode.")
		return
	}
	// wait for the main API is up and running.
	waitfor.API(fmt.Sprintf("%s%s", config.Cfg.Services[0].Url, readinessPath))
	mainAPI = config.Cfg.Services[0].Name
	// wait for the connection to establish.
	waitfor.True(messager.EnsureCanConnect, mainAPI)
	return mainAPI
}
