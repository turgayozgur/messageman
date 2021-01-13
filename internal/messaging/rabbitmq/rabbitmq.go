package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/turgayozgur/messageman/internal/metrics"
	"time"
)

var (
	connections map[string]*amqp.Connection
)

const (
	// RetryQueueNameSuffix constant.
	RetryQueueNameSuffix = "retry"
	// RetryQueueTTLMs constant.
	RetryQueueTTLMs = 30 * 1000
	// WaitToReconnectDuration constant.
	WaitToReconnectDuration = 5 * time.Second
	// DefaultConnectionName constant.
	DefaultConnectionName = "default"
)

// RabbitMQ messager
type RabbitMQ struct {
	exporter metrics.Exporter
	recover  chan string
}

// New ctor
func New(exporter metrics.Exporter) *RabbitMQ {
	connections = make(map[string]*amqp.Connection)
	r := &RabbitMQ{exporter: exporter}
	return r
}

// NotifyRecover .
func (r *RabbitMQ) NotifyRecover(ch chan string) chan string {
	r.recover = ch
	return r.recover
}

// Queue .
func (r *RabbitMQ) Queue(service string, name string, message []byte) (err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("could not be queued")
			r.exporter.IncSendError(service, name)
		}
		r.exporter.SendSeconds(time.Since(start), service, name)
	}()
	channel, err := r.channel(service, name, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := channel.Close(); err != nil {
			log.Error().Msgf("failed to close the channel. %v", err)
		}
	}()
	if err := r.exchange(channel, name); err != nil {
		return err
	}
	if err := r.send(channel, name, name, message); err != nil {
		return err
	}
	log.Debug().Str("name", name).Msgf("queued")
	return nil
}

// Work .
func (r *RabbitMQ) Work(service string, name string, callback func([]byte) bool) error {
	channel, err := r.channel(service, name, callback)
	if err != nil {
		return err
	}
	err = r.consume(channel, service, name, callback, false)
	if err != nil {
		return err
	}
	r.exporter.IncConsumer(service, name)
	return nil
}

// Publish .
func (r *RabbitMQ) Publish(service string, name string, message []byte) (err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("could not be published")
			r.exporter.IncSendError(service, name)
		}
		r.exporter.SendSeconds(time.Since(start), service, name)
	}()
	channel, err := r.channel(service, name, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := channel.Close(); err != nil {
			log.Error().Msgf("failed to close the channel. %v", err)
		}
	}()
	if err := r.exchange(channel, name); err != nil {
		return err
	}
	if err := r.send(channel, name, name, message); err != nil {
		return err
	}
	log.Debug().Str("name", name).Msgf("published")
	return nil
}

// Subscribe .
func (r *RabbitMQ) Subscribe(service string, name string, callback func([]byte) bool) error {
	channel, err := r.channel(service, name, callback)
	if err != nil {
		return err
	}
	err = r.consume(channel, service, name, callback, true)
	if err != nil {
		return err
	}
	r.exporter.IncConsumer(service, name)
	return nil
}

func (r *RabbitMQ) getQueueName(service string, name string, pubSub bool) string {
	var queueName string
	if pubSub {
		queueName = fmt.Sprintf("%s.%s", name, service)
	} else {
		queueName = name
	}
	return queueName
}

func (r *RabbitMQ) getRetryExchangeName(exchangeName string) string {
	return fmt.Sprintf("%s.%s", exchangeName, RetryQueueNameSuffix)
}
