package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "85"

type Config struct {
	Conn *amqp.Connection
}

func main() {
	conn, err := connectionAMPQ()
	if err != nil {
		log.Printf("Failed to connect rabbitmq: %s", err.Error())
		return
	}
	app := Config{
		Conn: conn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectionAMPQ() (*amqp.Connection, error) {
	var count int64
	var backOff = 1 * time.Second
	var conn *amqp.Connection
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
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
