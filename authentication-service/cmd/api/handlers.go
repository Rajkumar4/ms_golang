package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type requestPayload struct {
	Email    string `json:"email`
	Password string `json:"password"`
}

func (app *Config) authenticate(w http.ResponseWriter, r *http.Request) {
	reqPayload := &requestPayload{}
	err := app.readJSON(w, r, reqPayload)
	if err != nil {
		log.Printf("Failed to read json request data %s",err.Error())
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	
	user, err := app.Models.User.GetByEmail(reqPayload.Email)
	if err != nil {
		log.Printf("Failed to get user from database %s", err.Error())
		app.errorJSON(w, errors.New("invalid credaintials"), http.StatusBadRequest)
		return
	}
	valid, err := user.PasswordMatches(reqPayload.Password)
	if err != nil && !valid {
		log.Printf("Failed to match password")
		app.errorJSON(w, errors.New("invalid credaintials"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("logged in successfully %s", reqPayload.Email),
		Data:    user,
	}

	err = app.logRequest("event", "Authenticate request is succesfull")
	if err != nil {
		log.Printf("Failed to log request: %s", err.Error())
		app.writeJSON(w, http.StatusInternalServerError, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name string, data string) error {
	var logresuest struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	bytedata, _ := json.MarshalIndent(logresuest, "", "\t")
	loggerService := "http://logger-service:2001/logger"

	req, err := http.NewRequest("POST", loggerService, bytes.NewBuffer(bytedata))
	if err != nil {
		log.Println("Failed make request to logger service: %s", err.Error())
		return err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request a log %s", err.Error())
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		log.Printf("Invalid logger request")
		return errors.New("Invalid request")
	}
	return nil
}
