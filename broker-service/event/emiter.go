package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emiter struct {
	Conn *amqp.Connection
}

func (e *Emiter) setUp() error {
	channel, err := e.Conn.Channel()
	if err != nil {
		log.Printf("Failed to create a channel %s", err.Error())
		return err
	}

	defer channel.Close()
	decalreExchange(channel)
	return nil
}

func (e *Emiter) Push(event string, severity string) error {
	channel, err := e.Conn.Channel()
	if err != nil {
		log.Printf("Failed to create a channel %s", err.Error())
		return err
	}

	defer channel.Close()

	err = channel.Publish(
		"log_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		log.Printf("Failed to publish Data in rabbitmq: %s", err.Error())
		return err
	}
	return nil
}

func NewEmiter(conn *amqp.Connection) (Emiter, error) {
	emiter := Emiter{
		Conn: conn,
	}
	err := emiter.setUp()
	if err != nil {
		log.Printf("Failed to setup of emiter: %s", err.Error())
		return Emiter{}, err
	}
	return emiter, nil
}
