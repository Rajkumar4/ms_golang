package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const webPort = "2025"

type Config struct {
	Mailer Mail
}

func main() {
	app := Config{
		Mailer: getMailer(),
	}
	log.Printf("Starting mail service at port:%s\n", webPort)
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start server: %s", err.Error())
		return
	}
}

func getMailer() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		UserName:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		FromName:    os.Getenv("FROM_NAME"),
	}
	return m
}
