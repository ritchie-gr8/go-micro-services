package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type (
	RequestPayload struct {
		Action string      `json:"action"`
		Auth   AuthPayload `json:"auth,omitempty"`
		Log    LogPayload  `json:"log,omitempty"`
	}

	AuthPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	LogPayload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
)

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItem(w, requestPayload.Log)
	default:
		app.errorJSON(w, errors.New("unknow action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create json that will be sent to the auth service
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		app.errorJSON(w, errors.New("cannot marshal the payload"), http.StatusBadRequest)
	}

	// call the service
	authURL := "http://authentication-service/authenticate"
	request, err := http.NewRequest("POST",
		authURL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	// make sure we get back correct status code
	if resp.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if resp.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create response
	var jsonFromService jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Authenticated!",
		Data:    jsonFromService.Data,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// TODO: refactor a function
// (maybe create a helper func to reduce boilerplate code when sending back a response)
func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		app.errorJSON(w, errors.New("cannot marshal the payload"), http.StatusBadRequest)
	}

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Logged!",
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
