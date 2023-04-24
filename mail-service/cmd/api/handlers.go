package main

import (
	"fmt"
	"log"
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var reqPayLoad mailMessage
	err := app.readJSON(w, r, &reqPayLoad)
	if err != nil {
		log.Printf("Failed to read data from json file: %s", err.Error())
		return
	}
	log.Printf("Check jaosn input %v", reqPayLoad)
	msg := Message{
		From:    reqPayLoad.From,
		To:      reqPayLoad.To,
		Subject: reqPayLoad.Subject,
		Data:    reqPayLoad.Message,
	}
	log.Printf("check msg values in halndelr %v\n", msg.From)
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Printf("Failed to send mail: %s", err.Error())
		app.errorJSON(w, err, http.StatusForbidden)
		return
	}
	resp := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("mail is sent %s from %s", reqPayLoad.From, reqPayLoad.To),
	}
	app.writeJSON(w, http.StatusAccepted, resp)
}
