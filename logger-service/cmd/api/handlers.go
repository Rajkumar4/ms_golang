package main

import (
	"log"
	"logger/data"
	"net/http"

	"github.com/google/uuid"
)

type jsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLogs(w http.ResponseWriter, r *http.Request) {
	var reqPayLoad jsonPayload

	err := app.readJSON(w, r, &reqPayLoad)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, err)
		log.Printf("Failed to read data from: %s", err.Error())
		return
	}
	uid,err:=uuid.NewUUID()
	if err!=nil{
		log.Printf("Failed to genrate uid %s",err.Error())
		app.writeJSON(w,http.StatusConflict,err)
		return
	}
	logger := data.Logger{
		ID: uid.String(),
		Name: reqPayLoad.Name,
		Data: reqPayLoad.Data,
	}
	err = logger.Insert()
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, err)
		log.Printf("Failed to insert data: %s", err.Error())
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}
	app.writeJSON(w, http.StatusAccepted, resp)
}
