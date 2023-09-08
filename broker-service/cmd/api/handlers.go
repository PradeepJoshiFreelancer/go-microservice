package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action string         `json:"action"`
	Auth   AuthPayload    `json:"auth,omitempty"`
	Log    LogPayload     `json:"log,omitempty"`
	Mail   MailerPaylload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailerPaylload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Broker hit",
	}
	app.WriteJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmition(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItemViaRabbit(w, requestPayload.Log)
	case "mail":
		app.sendEmail(w, requestPayload.Mail)
	default:
		app.ErrorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	//create a request json
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	//call the service
	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	//check valid response
	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling logger service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged"

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	//create a request json
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	//call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	//check valid response
	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJson(w, errors.New("unauthrized user"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.ErrorJson(w, err, http.StatusUnauthorized)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) sendEmail(w http.ResponseWriter, msg MailerPaylload) {
	//create a request json
	jsonData, _ := json.MarshalIndent(msg, "", "\t")
	//call the service
	request, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
		app.ErrorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	//check valid response
	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling mailer service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Email Sent"

	app.WriteJson(w, http.StatusAccepted, payload)

}
func (app *Config) logItemViaRabbit(w http.ResponseWriter, entry LogPayload) {
	err := app.pushToQueue(entry.Name, entry.Data)
	if err != nil {
		log.Println(err)
		app.ErrorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via Rabbit MQ"
	app.WriteJson(w, http.StatusAccepted, payload)

}

func (app *Config) pushToQueue(name, msg string) error {
	emiter, err := event.NewEmiter(app.Rabbit)
	if err != nil {
		log.Println(err)
		return err
	}
	payload := LogPayload{
		Name: name,
		Data: msg,
	}
	log.Println("name", name)
	log.Println("msg", msg)
	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emiter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}
