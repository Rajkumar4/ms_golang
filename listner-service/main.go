package main

import (
	"errors"
	"fmt"
	"listner/event"
	"log"
	"math"
	"time"

	ampq "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := connectionAMPQ()
	if err != nil {
		log.Printf("Failed to connect rabbitMQ: %s", err.Error())
		return 
	}
	defer conn.Close()

	consumer, err := event.NewConsumer(conn)
	if err != nil {
		log.Printf("Failed to create a new consumer %s", err.Error())
		return
	}
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Printf("failed to listen %s", err.Error())
		return
	}
}

func connectionAMPQ() (*ampq.Connection, error) {
	var count int64
	var backOff = 1 * time.Second
	var conn *ampq.Connection
	for {
		c, err := ampq.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("rabbitmq is not ready...")
			count++
		} else {
			log.Printf("Connected to rabbitmq")
			conn = c
			return conn, nil
		}
		if count < 5 {
			backOff = time.Duration(math.Pow(float64(count), 2))
			time.Sleep(backOff)
			log.Println("Retry for connectioin")
		} else {
			log.Println("Failed to connect rabbitmq")
			return nil, errors.New(fmt.Sprintf("quit after maximum try: %s", err.Error()))
		}
	}
}
