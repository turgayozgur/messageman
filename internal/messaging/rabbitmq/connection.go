package rabbitmq

import (
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/turgayozgur/messageman/config"
	"strings"
	"time"
)

// EnsureCanConnect to ensure connection is available to rabbitmq
func (r *RabbitMQ) EnsureCanConnect() (result bool) {
	url := config.Cfg.RabbitMQ.Url
	host := ""
	urlParts := strings.Split(url, "@")
	if len(urlParts) > 0 {
		host = urlParts[1]
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("failed to connect to the RabbitMQ server. host:%s error:%v", host, r)
			result = false
		}
	}()

	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}
	if err = conn.Close(); err != nil {
		panic(err)
	}
	log.Info().Msg("successfully connected to the RabbitMQ server")
	return true
}

func (r *RabbitMQ) connection(name string) *amqp.Connection {
	if name == "" {
		name = DefaultConnectionName
	}

	var connection *amqp.Connection
	if connection, ok := connections[name]; ok {
		return connection
	}

	url := config.Cfg.RabbitMQ.Url
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}
	connections[name] = conn
	r.exporter.IncConnection(name)

	go func() {
		for {
			reason, ok := <-connection.NotifyClose(make(chan *amqp.Error))
			r.exporter.DecConnection(name)
			// exit this goroutine if closed by developer
			if !ok {
				log.Info().Str("name", name).Msg("connection closed")
				break
			}
			log.Error().Str("name", name).Msgf("connection closed, reason: %v", reason)
			// reconnect if not closed by developer
			for {
				time.Sleep(WaitToReconnectDuration)
				conn, err := amqp.Dial(url)
				if err == nil {
					r.exporter.IncConnection(name)
					connection = conn
					connections[name] = conn
					log.Info().Str("name", name).Msg("successfully reconnected")
					r.recover <- name
					break
				}
				r.exporter.IncError("connection")
				log.Error().Err(err).Str("name", name).Msg("failed to reconnect")
			}
		}
	}()

	connection = conn
	return connection
}
