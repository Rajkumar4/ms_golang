package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string        `json:"action"`
	Auth   AuthPayload   `json:"auth",omitempty`
	Log    LoggerPayload `json:"log",omitempty`
	Mail   MailPayload   `json:"mail",omitempty`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type LoggerPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) authenticateSubmition(w http.ResponseWriter, r *http.Request) {
	var reqPayload = &RequestPayload{}

	err := app.readJSON(w, r, reqPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("Check the action values: %s", reqPayload.Action)
	switch reqPayload.Action {
	case "auth":
		app.authenticate(w, reqPayload.Auth)
	case "log":
		app.logViaRpc(w, reqPayload.Log)
	case "mail":
		app.sendMail(w, reqPayload.Mail)

	default:
		app.errorJSON(w, errors.New("invalid action"))

	}
}

func (app *Config) Logger(w http.ResponseWriter, entry LoggerPayload) {
	bytePayload, _ := json.MarshalIndent(entry, "", "\t")
	log.Printf("Check the value parse in re1 %v", entry)
	url := "http://logger-service:2001/logger"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytePayload))
	if err != nil {
		log.Printf("Failed to create a request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request to logger service: %s ", err.Error())
		return
	}
	if resp.StatusCode != http.StatusAccepted {
		app.writeJSON(w, http.StatusBadRequest, errors.New("Bad request"))
		log.Println("Failed request bad request")
		return
	}

	respBody := &jsonResponse{
		Error:   false,
		Message: "logging is complete",
	}

	app.writeJSON(w, http.StatusAccepted, respBody)
}

func (app *Config) authenticate(w http.ResponseWriter, auth AuthPayload) {
	bytePayload, _ := json.MarshalIndent(auth, "", "\t")
	req, err := http.NewRequest("POST", "http://authentication-service:8089/auth",
		bytes.NewBuffer(bytePayload))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request %s", err.Error())
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		log.Printf("Failed to make request response is not in order")
		app.errorJSON(w, errors.New("invaild credentials"), http.StatusBadRequest)
		return
	}

	var readResponseBody jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&readResponseBody)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, readResponseBody)
}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	jsonData, _ := json.MarshalIndent(m, "", "\t")
	mailURL := "http://mail-service:2025/mail"
	req, err := http.NewRequest("POST", mailURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create new request: %s", err.Error())
		app.errorJSON(w, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request to mailer service: %s", err.Error())
		app.errorJSON(w, err)
		return
	}
	if resp.StatusCode != http.StatusAccepted {
		log.Println("invalid request")
		app.errorJSON(w, errors.New("invalid request"))
		return
	}

	var jsonResp = jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("mail is send to %s  from %s", m.To, m.From),
	}
	app.writeJSON(w, http.StatusAccepted, jsonResp)
}

func (app *Config) logItemViaQueue(w http.ResponseWriter, payload LoggerPayload) {
	err := app.pushToQueue(payload.Name, payload.Data)
	if err != nil {
		log.Printf("Failed to push in queue in handler %s", err.Error())
		return
	}
	var resp jsonResponse = jsonResponse{
		Error:   false,
		Message: "Logged message using messging queue",
	}
	app.writeJSON(w, http.StatusAccepted, resp)
}

func (app *Config) pushToQueue(name, msg string) error {
	emiter, err := event.NewEmiter(app.Conn)
	if err != nil {
		log.Printf("Failed to create new emiter %s", err.Error())
		return err
	}
	lPayload := LoggerPayload{
		Name: name,
		Data: msg,
	}
	byteData, _ := json.Marshal(lPayload)
	err = emiter.Push(string(byteData), "Log_INFO")
	if err != nil {
		log.Printf("Failed to psuh data %s", err.Error())
		return err
	}
	return nil
}

type rpcPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRpc(w http.ResponseWriter, l LoggerPayload) {
	client, err := rpc.Dial("tcp", fmt.Sprintf("logger-service:5001"))
	if err != nil {
		log.Printf("Failed to dial rpc server %s", err.Error())
		app.errorJSON(w, err, http.StatusBadGateway)
		return
	}

	rpcPayload := rpcPayload{
		Name: l.Name,
		Data: l.Data,
	}
	log.Printf("check input json %v",rpcPayload)
	var resp *string
	err = client.Call("RpcServer.LogInfo", rpcPayload, &resp)
	if err != nil {
		log.Printf("Failed to get call the rpc function %s", err.Error())
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	jsonresp := jsonResponse{
		Error:   false,
		Message: "logged using rpc",
		Data:    resp,
	}
	app.writeJSON(w, http.StatusOK, jsonresp)
}
