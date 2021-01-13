package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"time"
)

func (r *RabbitMQ) channel(service string, name string, callback func([]byte) bool) (channel *amqp.Channel, err error) {
	connection := r.connection(name)
	channel, err = connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel. %v", err)
	}
	go func() {
		for {
			reason, ok := <-channel.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok {
				break
			}
			r.exporter.DecConsumer(service, name)
			log.Error().Msgf("channel closed, reason: %v", reason)
			// reconnect if not closed by developer
			for {
				time.Sleep(WaitToReconnectDuration)
				if connection.IsClosed() {
					continue
				}
				ch, err := connection.Channel()
				if err == nil {
					channel = ch
					if err := r.consume(channel, service, name, callback, false); err != nil {
						log.Error().Msgf("failed to recover consumer %s. %v", name, err)
						continue
					}
					r.exporter.IncConsumer(service, name)
					log.Info().Msgf("consumer %s successfully recovered.", name)
					break
				}
				r.exporter.IncError(service)
				log.Error().Err(err).Msg("failed to recreate channel")
			}
		}
	}()
	return channel, nil
}
