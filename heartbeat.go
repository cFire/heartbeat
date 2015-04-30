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
	Id			int
	HostId		int
	Name		string
	Port		int
	Ssl			bool
	Request		string
	Response	string
	Result		string
}

/* Host definition */
type Host struct {
	Id			int
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

/* Global synchronization object */
//var wg sync.WaitGroup

/* Global counter for processes and services */
var ncpu = 0

/* Global channel for result collection */
var ch = make(chan Service)

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
	for _,h := range conf.Hosts {
		ncpu += len(h.Services)
	}
	runtime.GOMAXPROCS(ncpu)
}

/* Master function for running checks */
func runChecks() {
	var sId = 0
	var hId = 0
	for _,h := range conf.Hosts {
		h.Id = hId
		for _,s := range h.Services {
			s.Id = sId
			s.HostId = hId
			go doCheck(s)
			sId++
		}
		hId++
	}
}

/* Slave function for running checks */
func doCheck(s Service) {
	s.Result = "All is well."
	ch <- s
}

/* Read results from channel */
func collectResults() []Service {
	var resultSet []Service
	var c = 0
	for c < ncpu {
		result := <-ch
		resultSet = append(resultSet, result)
		c++
	}

	return resultSet
}

/* Handler function for HTTP requests */
func handler(w http.ResponseWriter, r *http.Request) {
	/* Start checks */
	runChecks()

	/* Wait for checks to finish */
	var resultSet []Service
	resultSet = collectResults()

	/* Display results */
	for _,r := range resultSet {
		fmt.Print(r.Id)
		fmt.Print(" ")
		fmt.Print(r.Name)
		fmt.Print(" ")
		fmt.Println(r.Result)
	}
	var sId = 0
	var hId = 0
	for _,h := range conf.Hosts {
		for range h.Services {
			for _,r := range resultSet {
				if r.HostId == hId && r.Id == sId {
					fmt.Print(r.Id)
					fmt.Print(" ")
					fmt.Print(r.Name)
					fmt.Print(" ")
					fmt.Println(r.Result)
				}
			}
			sId++
		}
		hId++
	}

	fmt.Println("Done.")
	fmt.Fprintf(w, "All done!")
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