package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"time"
)

// Publish
func (r *RabbitMQ) Publish(publisher string, eventName string, message []byte) (err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Message could not be published")
			r.exporter.IncPublishError(publisher, eventName)
		}
		r.exporter.PublishSeconds(time.Since(start), publisher, eventName)
	}()
	channel, err := r.channel(publisher, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := channel.Close(); err != nil {
			log.Error().Msgf("Failed to close the channel. %v", err)
		}
	}()
	if err := r.exchangeForPubSub(channel, eventName); err != nil {
		return err
	}
	if err := r.publish(channel, eventName, message); err != nil {
		return err
	}
	log.Debug().Str("event", eventName).Msgf("Message published")
	return nil
}

func (r *RabbitMQ) Subscribe(subscriber string, eventName string, fn func([]byte) bool) error {
	channel, err := r.channel(subscriber, &Consumer{fn: fn, name: eventName})
	if err != nil {
		return err
	}
	queueName := r.getPubSubQueueName(subscriber, eventName)
	if err := r.prepareChannelForSubscribe(channel, queueName, eventName, subscriber); err != nil {
		return err
	}
	if err := r.prepareChannelForRetry(channel, r.getPubSubQueueName(subscriber, eventName), eventName); err != nil {
		return err
	}
	msgs, err := channel.Consume(
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
		for d := range msgs {
			start := time.Now()
			body := d.Body
			success := r.invokeConsumerFunc(body, fn)
			if !success {
				if err := r.send(channel, r.getRetryExchangeName(eventName), queueName, queueName, body); err != nil {
					log.Error().Msgf("Failed to publish to retry queue. %v", err)
				}
				r.exporter.IncHandleError(subscriber, eventName)
			}
			if err := d.Ack(false); err != nil {
				log.Error().Msgf("Failed to send ACK. %v", err)
			}
			r.exporter.HandleSeconds(time.Since(start), subscriber, eventName)
		}
	}()
	r.exporter.IncSubscriber(subscriber, eventName)
	return nil
}

func (r *RabbitMQ) publish(channel *amqp.Channel, eventName string, message []byte) error {
	err := channel.Publish(
		eventName, // exchange
		eventName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message. %v", err)
	}
	return nil
}

func (r *RabbitMQ) exchangeForPubSub(channel *amqp.Channel, eventName string) error {
	err := channel.ExchangeDeclare(
		eventName, // name
		"direct",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a main exchange. %v", err)
	}
	return nil
}

func (r *RabbitMQ) prepareChannelForSubscribe(channel *amqp.Channel, queueName string, eventName string, subscriber string) error {
	if err := r.exchangeForPubSub(channel, eventName); err != nil {
		return err
	}

	_, err := channel.QueueDeclare(
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
		eventName, // routing key
		eventName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange. %v", err)
	}
	err = channel.QueueBind(
		queueName, // queue name
		queueName, // routing key to retry for specific subscriber
		eventName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange. %v", err)
	}
	return nil
}

func (r *RabbitMQ) getPubSubQueueName(subscriber string, eventName string) string {
	return fmt.Sprintf("%s.%s", subscriber, eventName)
}
