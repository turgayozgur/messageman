package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

func (r *RabbitMQ) send(channel *amqp.Channel, name string, route string, message []byte) error {
	err := channel.Publish(
		name,  // exchange
		route, // routing key
		false, // mandatory
		false, // immediate
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
