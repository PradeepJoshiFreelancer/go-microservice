package main

import (
	"log-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload
	_ = app.readJSON(w, r, &requestPayload)

	//insert the data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	err := app.Models.LogEntry.InsertLogEntry(event)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	rspn := jsonResponse{
		Error:   false,
		Message: "logged",
	}
	app.WriteJson(w, http.StatusAccepted, rspn)
}
