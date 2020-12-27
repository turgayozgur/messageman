package rabbitmq

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/turgayozgur/messageman/config"
	"github.com/turgayozgur/messageman/internal/metrics"
)

var (
	connections map[string]*amqp.Connection
)

const (
	// MainExchangeName constant.
	MainExchangeName = "messsager.exchange.main"
	// MainRetryExchangeName constant.
	MainRetryExchangeName = "messsager.exchange.main.retry"
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
}

// New ctor
func New(exporter metrics.Exporter) *RabbitMQ {
	connections = make(map[string]*amqp.Connection)
	r := &RabbitMQ{exporter: exporter}
	return r
}

// Consumer .
type Consumer struct {
	name string
	fn   func([]byte) bool
}

// EnsureCanConnect to ensure connection is available to rabbitmq
func (r *RabbitMQ) EnsureCanConnect(params []interface{}) (result bool) {
	mainAPI := fmt.Sprintf("%v", params[0])
	url := config.Cfg.RabbitMQ.Url
	host := ""
	urlParts := strings.Split(url, "@")
	if len(urlParts) > 0 {
		host = urlParts[1]
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error().Str("mainApi", mainAPI).
				Msgf("Failed to connect to the RabbitMQ server. host:%s error:%v", host, r)
			result = false
		}
	}()

	r.connection(mainAPI)
	log.Info().Str("mainApi", mainAPI).Msg("Successfully connected to the RabbitMQ server")
	return true
}

// Send to queue
func (r *RabbitMQ) Send(mainAPI string, name string, message []byte) (err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Job could not be sent")
			r.exporter.IncSendError(mainAPI, name)
		}
		r.exporter.SendSeconds(time.Since(start), mainAPI, name)
	}()
	channel, err := r.channel(mainAPI, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := channel.Close(); err != nil {
			log.Error().Msgf("Failed to close the channel. %v", err)
		}
	}()
	if err := r.prepareChannel(channel, name); err != nil {
		return err
	}
	if err := r.publish(channel, MainExchangeName, name, message); err != nil {
		return err
	}
	log.Debug().Str("queue", name).Msgf("Job sent")
	return nil
}

// Receive messages from message queue
func (r *RabbitMQ) Receive(mainAPI string, name string, fn func([]byte) bool) error {
	channel, err := r.channel(mainAPI, &Consumer{fn: fn, name: name})
	if err != nil {
		return err
	}
	err = r.receive(channel, mainAPI, name, fn)
	if err != nil {
		return err
	}
	r.exporter.IncWorker(mainAPI, name)
	return nil
}

func (r *RabbitMQ) receive(channel *amqp.Channel, mainAPI string, name string, fn func([]byte) bool) error {
	if err := r.prepareChannel(channel, name); err != nil {
		return err
	}
	if err := r.prepareChannelForRetry(channel, name, MainExchangeName); err != nil {
		return err
	}

	msgs, err := channel.Consume(
		name,  // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			start := time.Now()
			body := d.Body
			success := r.invokeConsumerFunc(body, fn)
			if !success {
				if err := r.publish(channel, MainRetryExchangeName, name, body); err != nil {
					log.Error().Msgf("Failed to publish to retry queue. %v", err)
				}
				r.exporter.IncReceiveError(mainAPI, name)
			}
			if err := d.Ack(false); err != nil {
				log.Error().Msgf("Failed to send ACK. %v", err)
			}
			r.exporter.ReceiveSeconds(time.Since(start), mainAPI, name)
		}
	}()

	return nil
}

func (r *RabbitMQ) connection(mainAPI string) *amqp.Connection {
	if mainAPI == "" {
		mainAPI = DefaultConnectionName
	}

	var connection *amqp.Connection
	if connection, ok := connections[mainAPI]; ok {
		return connection
	}

	url := config.Cfg.RabbitMQ.Url
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}
	connections[mainAPI] = conn
	r.exporter.IncConnection(mainAPI)

	go func() {
		for {
			reason, ok := <-connection.NotifyClose(make(chan *amqp.Error))
			r.exporter.DecConnection(mainAPI)
			// exit this goroutine if closed by developer
			if !ok {
				log.Info().Msg("Connection closed")
				break
			}
			log.Error().Msgf("Connection closed, reason: %v", reason)
			// reconnect if not closed by developer
			for {
				time.Sleep(WaitToReconnectDuration)
				conn, err := amqp.Dial(url)
				if err == nil {
					r.exporter.IncConnection(mainAPI)
					connection = conn
					log.Info().Msg("Successfully reconnected")
					break
				}
				r.exporter.IncError("connection")
				log.Error().Err(err).Msg("Failed to reconnect")
			}
		}
	}()

	connection = conn
	return connection
}

func (r *RabbitMQ) channel(mainAPI string, consumer *Consumer) (channel *amqp.Channel, err error) {
	connection := r.connection(mainAPI)
	channel, err = connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel. %v", err)
	}
	go func() {
		for {
			reason, ok := <-channel.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok || consumer == nil {
				break
			}
			r.exporter.DecWorker(mainAPI, consumer.name)
			log.Error().Msgf("Channel closed, reason: %v", reason)
			// reconnect if not closed by developer
			for {
				time.Sleep(WaitToReconnectDuration)
				if connection.IsClosed() {
					continue
				}
				ch, err := connection.Channel()
				if err == nil {
					channel = ch
					if err := r.receive(channel, mainAPI, consumer.name, consumer.fn); err != nil {
						log.Error().Msgf("Failed to recover consumer %s. %v", consumer.name, err)
						continue
					}
					r.exporter.IncWorker(mainAPI, consumer.name)
					log.Info().Msgf("Consumer %s successfully recovered.", consumer.name)
					break
				}
				r.exporter.IncError(mainAPI)
				log.Error().Err(err).Msg("Failed to recreate channel")
			}
		}
	}()
	return channel, nil
}

func (r *RabbitMQ) prepareChannel(channel *amqp.Channel, queueName string) error {
	err := channel.ExchangeDeclare(
		MainExchangeName, // name
		"direct",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a main exchange. %v", err)
	}

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue. %v", err)
	}

	err = channel.QueueBind(
		queueName,        // queue name
		queueName,        // routing key
		MainExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to bind queue to exchange. %v", err)
	}
	return nil
}

func (r *RabbitMQ) prepareChannelForRetry(channel *amqp.Channel, queueName string, mainExchangeName string) error {
	err := channel.ExchangeDeclare(
		MainRetryExchangeName, // name
		"direct",              // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a retry exchange. %v", err)
	}

	name := fmt.Sprintf("%s.%s", queueName, RetryQueueNameSuffix)

	_, err = channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": mainExchangeName,
			"x-message-ttl":          RetryQueueTTLMs,
		}, // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a retry queue. %v", err)
	}

	err = channel.QueueBind(
		name,                  // queue name
		queueName,             // routing key
		MainRetryExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to bind retry queue to exchange. %v", err)
	}
	return nil
}

func (r *RabbitMQ) publish(channel *amqp.Channel, exchangeName string, queueName string, message []byte) error {
	err := channel.Publish(
		exchangeName, // exchange
		queueName,    // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		return fmt.Errorf("Failed to publish a message. %v", err)
	}
	return nil
}

func (r *RabbitMQ) invokeConsumerFunc(body []byte, fn func([]byte) bool) (result bool) {
	defer func() {
		if rc := recover(); rc != nil {
			log.Error().Msgf("Error when invoking the consumer function: %+v", rc)
			result = false
		}
	}()
	return fn(body)
}
