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

// relay This is the main webhook handler function
func relay(w http.ResponseWriter, r *http.Request) {
	var newMsg []MsgIn
	// lookup MS Teams target webhook URL
	target, found := config.Targets[mux.Vars(r)["target"]]
	if !(found) {
		log.Printf("Unknown alert target: %v\n", mux.Vars(r)["target"])
		return
	}

	// read the contents of the Rancher webhook post
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading post body")
		return
	}

	// place the Rancher webhook contents into the log
	log.Printf("Body: %v", string(reqBody))

	// unmarshal the rancher webhook JSON
	json.Unmarshal(reqBody, &newMsg)

	// loop over each alert from Ranchar and post them as individual messages in MS Teams
	for _, msg := range newMsg {
		sendMsg(msg, target)
	}

	//	Tell Rancher we posted the messages
	w.WriteHeader(http.StatusCreated)
}

// sendMsg convert each alert to simple text format and post to MS Teams webhook
func sendMsg(msg MsgIn, target string) {
	// simplest MS Teams webhook format is a single text field
	msgOut := "{\"text\": \""

	// loop over the rancher labels map and append each entry to the text field
	for k, v := range msg.Labels {
		msgOut = msgOut + k + "=" + v + "<br />"
	}

	// loop over the rancher annotations map and append each entry to the text field
	for k, v := range msg.Annotations {
		msgOut = msgOut + k + "=" + v + "<br />"
	}

	// add the timestamp and generatorURL fields to the end of the text message
	msgOut = msgOut + "startsAt=" + msg.StartsAt + "<br />"
	msgOut = msgOut + "endsAt=" + msg.EndsAt + "<br />"
	msgOut = msgOut + "generatorURL=" + msg.GeneratorURL + "<br />"

	// close out the json structure
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

	// create a https post request for MS Teams
	request, err := http.NewRequest(http.MethodPost, target, bytes.NewBuffer([]byte(msgOut)))
	if err != nil {
		log.Printf("Error while creating request to post alert to MS Teams: %v\n", err.Error())
		return
	}

	// tell MS Teams that this is JSON content
	request.Header.Set("Content-Type", "application/json")

	// post the request to MS Teams
	response, err := netClient.Do(request)
	if err != nil {
		log.Printf("Error while communicating with MS Teams: %v\n", err.Error())
		return
	}

	// read the response from MS Teams and log it
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	log.Printf("response code: %v\n", response.StatusCode)
	log.Printf("response body: %v\n", string(body))
}
