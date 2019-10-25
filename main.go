package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krenaut1/goconfig"
)

// Config this structure defines the application properties
type Config struct {
	ServerAddr string
	ServerPort int
	Targets    map[string]string
}

var config = Config{}

func main() {

	// load app properties from config file based on value of PROFILE env variable
	loadAppProperites()

	// map all supported request
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/relay/{target}", relay).Methods("POST")
	listenAddrPort := fmt.Sprintf("%v:%v", config.ServerAddr, config.ServerPort)
	log.Printf("listening on: %v:%v", config.ServerAddr, config.ServerPort)
	// start listening for requests
	log.Fatal(http.ListenAndServe(listenAddrPort, router))
}

func loadAppProperites() {
	// populate my configstructure from ./config directory using PROFILE environment variable
	err := goconfig.GoConfig(&config) 
	if err != nil {
		log.Fatalf("Error loading application properties: %v", err.Error())
	}
}
