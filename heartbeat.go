package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"io"
	"os"
	"runtime"
)

/* Service definition */
type Service struct {
	Name		string
	Port		int
	Ssl			bool
	Request		string
	Response	string
}

/* Host definition */
type Host struct {
	Name		string
	Address		string
	Services	[]Service
}

/* Config object definition */
type Config struct {
	ListenAddress	string
	Hosts			[]Host
}

/* Global config object */
var conf Config

/* Load configuration */
func loadConfig() {
	/* Open config file */
	input, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}

	/* Parse json into Config type */
	dec := json.NewDecoder(input)
	for {
		err := dec.Decode(&conf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
	}
}

/* Count and set the number of threads we need */
func setThreading() {
	var ncpu = 0

	for _,h := range conf.Hosts {
		ncpu += len(h.Services)
	}
	runtime.GOMAXPROCS(ncpu)
}

/* Master function for running checks */
func runChecks() {
	var id = 0
	for _,h := range conf.Hosts {
		for _,s := range h.Services {
			go doCheck(id, s)
			id++
		}
	}
}

/* Slave function for running checks */
func doCheck(id int, s Service) {
	fmt.Println(id, s.Name)
}

/* Handler function for HTTP requests */
func handler(w http.ResponseWriter, r *http.Request) {
	/* Start checks */
	runChecks()
	fmt.Fprintf(w, "%s", conf)
}

func main() {
	/* Setup */
	loadConfig()
	setThreading()

	/* Create HTTP handler function */
	http.HandleFunc("/", handler)

	/* Listen port */
	http.ListenAndServe(conf.ListenAddress, nil)
}