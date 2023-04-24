package event

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	ampq "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Conn      *ampq.Connection
	QueueName string
}

func NewConsumer(conn *ampq.Connection) (*Consumer, error) {
	consumer := &Consumer{
		Conn: conn,
	}

	err := consumer.setUp()
	if err != nil {
		log.Printf("Failed to set up consumer: %s", err.Error())
		return nil, err
	}
	return consumer, nil
}

func (c *Consumer) setUp() error {
	channel, err := c.Conn.Channel()
	if err != nil {
		log.Printf("Failed to get channel %s", err.Error())
		return err
	}
	return decalreExchange(*channel)
}

type PayLoad struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Consumer) Listen(topics []string) error {
	ch, err := c.Conn.Channel()
	if err != nil {
		log.Printf("Failed to get channel: %s", err.Error())
		return err
	}
	defer ch.Close()

	queue, err := delareQueue(*ch)
	if err != nil {
		log.Printf("failed to delare queue: %s", err.Error())
		return nil
	}

	for _, v := range topics {
		err = ch.QueueBind(queue.Name, v, "log_topic", false, nil)
		if err != nil {
			log.Printf("Failed to bind topic to queue: %s", err.Error())
			return err
		}
	}
	message, err := ch.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("failed to consume message: %s", err.Error())
		return err
	}

	forever := make(chan bool)

	go func() {
		for m := range message {
			var payload PayLoad
			_ = json.Unmarshal(m.Body, &payload)
			go handlePayLoad(payload)
		}
	}()

	log.Printf("Waiting for message on [Exchange,Queue],[log_topic,%s]", queue.Name)
	<-forever
	return nil
}

func handlePayLoad(payload PayLoad) error {
	switch payload.Name {
	case "log", "event":
		if err := logEvent(payload); err != nil {
			log.Printf("Failed to log entry: %s", err.Error())
			return err
		}
	default:
		if err := logEvent(payload); err != nil {
			log.Printf("Failed to log entry: %s", err.Error())
			return err
		}
	}
	return nil
}

func logEvent(entry PayLoad) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	URL := "http://logger-service/log"
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create http request %s", err.Error())
		return nil
	}

	req.Header.Set("Content/Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to get resp of http request %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Printf("Failed to request in logger service %s", err.Error())
		return err
	}
	return nil
}
