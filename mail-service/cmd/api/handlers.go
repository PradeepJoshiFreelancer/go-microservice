package main

import (
	"log"
	"net/http"
)

func (app *Config) SendEmail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var resquestPayload mailMessage

	err := app.readJson(w, r, &resquestPayload)
	if err != nil {
		log.Println(err)

		app.ErrorJson(w, err)
		return
	}

	msg := Message{
		From:    resquestPayload.From,
		To:      resquestPayload.To,
		Subject: resquestPayload.Subject,
		Data:    resquestPayload.Message,
	}

	err = app.Mailer.SentSMTMessage(msg)
	if err != nil {
		log.Println(err)

		app.ErrorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "sent to " + resquestPayload.To,
	}
	app.WriteJson(w, http.StatusAccepted, payload)
}
