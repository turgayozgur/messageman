package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"time"
)

func (r *RabbitMQ) consume(channel *amqp.Channel, service string, name string, callback func([]byte) bool, pubSub bool) error {
	if err := r.bind(channel, service, name, pubSub); err != nil {
		return err
	}
	if err := r.bindRetry(channel, service, name, pubSub); err != nil {
		return err
	}

	queueName := r.getQueueName(service, name, pubSub)

	messages, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	go func() {
	loop:
		for {
			select {
			case d := <-messages:
				start := time.Now()
				body := d.Body
				success := r.invokeConsumerFunc(body, callback)
				if !success {
					if err := r.send(channel, r.getRetryExchangeName(name), queueName, body); err != nil {
						log.Error().Msgf("failed to publish to retry queue. %v", err)
					}
					r.exporter.IncConsumeError(service, name)
				}
				if err := d.Ack(false); err != nil {
					log.Error().Msgf("failed to send ACK. %v", err)
				}
				r.exporter.ConsumeSeconds(time.Since(start), service, name)
			case <-channel.NotifyClose(make(chan *amqp.Error)):
				log.Error().Str("service", service).Str("name", name).Msgf("consumer stopped. %v", err)
				break loop
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) exchange(channel *amqp.Channel, name string) error {
	return channel.ExchangeDeclare(
		name,     // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}

func (r *RabbitMQ) bind(channel *amqp.Channel, service string, name string, pubSub bool) error {
	err := r.exchange(channel, name)
	if err != nil {
		return fmt.Errorf("failed to declare a main exchange. %v", err)
	}

	queueName := r.getQueueName(service, name, pubSub)

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue. %v", err)
	}

	err = channel.QueueBind(
		queueName, // queue name
		name,      // routing key
		name,      // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange. %v", err)
	}

	err = channel.QueueBind(
		queueName, // queue name
		queueName, // routing key to retry for specific consumer
		name,      // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange. %v", err)
	}

	return nil
}

func (r *RabbitMQ) bindRetry(channel *amqp.Channel, service string, name string, pubSub bool) error {
	retryExchangeName := r.getRetryExchangeName(name)
	err := channel.ExchangeDeclare(
		retryExchangeName, // name
		"direct",          // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a retry exchange. %v", err)
	}

	queueName := r.getQueueName(service, name, pubSub)
	retryQueueName := fmt.Sprintf("%s.%s", queueName, RetryQueueNameSuffix)

	_, err = channel.QueueDeclare(
		retryQueueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		amqp.Table{
			"x-dead-letter-exchange": name,
			"x-message-ttl":          RetryQueueTTLMs,
		}, // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a retry queue. %v", err)
	}

	err = channel.QueueBind(
		retryQueueName,    // queue name
		queueName,         // routing key
		retryExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind retry queue to exchange. %v", err)
	}
	return nil
}

func (r *RabbitMQ) invokeConsumerFunc(body []byte, callback func([]byte) bool) (result bool) {
	defer func() {
		if rc := recover(); rc != nil {
			log.Error().Msgf("error when invoking the consumer function: %+v", rc)
			result = false
		}
	}()
	if callback == nil {
		log.Warn().Msgf("callback function is nil. Make sure you have correct setup for your consumer")
		return false
	}
	return callback(body)
}
