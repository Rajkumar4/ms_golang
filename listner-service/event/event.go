package event

import ampq "github.com/rabbitmq/amqp091-go"

func decalreExchange(ch ampq.Channel) error {
	return ch.ExchangeDeclare(
		"log_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func delareQueue(ch ampq.Channel) (ampq.Queue, error) {
	return ch.QueueDeclare(
		"",
		true,
		false,
		true,
		false,
		nil)
}
