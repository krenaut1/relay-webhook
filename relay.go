package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// MsgIn describe the rancher webhook format which is actually an array of this structure
type MsgIn struct {
	Labels       map[string]string
	Annotations  map[string]string
	StartsAt     string
	EndsAt       string
	GeneratorURL string
}

func relay(w http.ResponseWriter, r *http.Request) {
	var newMsg []MsgIn
	target := mux.Vars(r)["target"]
	if !(target == "dev" || target == "prod") {
		log.Printf("Unknown alert target: %v\n", target)
		return
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading post body")
		return
	}
	log.Printf("Body: %v", string(reqBody))

	json.Unmarshal(reqBody, &newMsg)
	for _, msg := range newMsg {
		sendMsg(msg, target)
	}
	//	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)
	//	json.NewEncoder(w).Encode(newEvent)
}

func sendMsg(msg MsgIn, target string) {
	url := ""
	msgOut := "{\"text\": \""
	for k, v := range msg.Labels {
		msgOut = msgOut + k + "=" + v + "<br />"
	}
	for k, v := range msg.Annotations {
		msgOut = msgOut + k + "=" + v + "<br />"
	}
	msgOut = msgOut + "startsAt=" + msg.StartsAt + "<br />"
	msgOut = msgOut + "endsAt=" + msg.EndsAt + "<br />"
	msgOut = msgOut + "generatorURL=" + msg.GeneratorURL + "<br />"
	msgOut = msgOut + "\"}"

	// configure a network transport object with timeout options
	// for the http client that we are about to create
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
	}

	// create a http client with timeout settings and the network transport
	// settings that we just defined in netTransport
	var netClient = &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}

	// create an https post for MS Teams
	if target == "dev" {
		url = "https://outlook.office.com/webhook/1dd1f352-0db5-44a1-bef5-d7be1fa950c7@2567b4c1-b0ed-40f5-aee3-58d7c5f3e2b2/IncomingWebhook/0b3167c38d114499be6cf8df5e0229b0/a8adaf56-d56d-451c-864d-3e36c366a366"
	}
	if target == "prod" {
		url = "https://outlook.office.com/webhook/1dd1f352-0db5-44a1-bef5-d7be1fa950c7@2567b4c1-b0ed-40f5-aee3-58d7c5f3e2b2/IncomingWebhook/5cc384a9c26c4048a531851eafc0ce3c/a8adaf56-d56d-451c-864d-3e36c366a366"
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(msgOut)))
	if err != nil {
		log.Printf("Error while creating request to post alert to MS Teams: %v\n", err.Error())
		return
	}

	request.Header.Set("Content-Type", "application/json")

	// post the request to teams
	response, err := netClient.Do(request)
	if err != nil {
		log.Printf("Error while communicating with MS Teams: %v\n", err.Error())
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	log.Printf("response code: %v\n", response.StatusCode)
	log.Printf("response body: %v\n", string(body))
}
