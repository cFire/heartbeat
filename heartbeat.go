package main

import (
	"fmt"
	"net"
	"time"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"log"
	"io"
	"os"
	"runtime"
	"strings"
	"strconv"
)

/* Service definition */
type Service struct {
	Id			int
	HostId		int
	Name		string
	Host		string
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
			s.Host = h.Address
			go doCheck(s)
			sId++
		}
		hId++
	}
}

/* Slave function for running checks */
func doCheck(s Service) {
	if s.Ssl == true {
		conn, err := tls.Dial("tcp", s.Host + ":" + strconv.Itoa(s.Port), &tls.Config{})
		if err != nil {
			s.Result = "<span style='color:red'>FAILED: " + err.Error() + "</span>"
		} else {
			if s.Request != "" {
				fmt.Fprintf(conn, s.Request)
			}

			var r = make([]byte, len(s.Response))
			_,e := conn.Read(r)
			if e != nil {
				if e != io.EOF {
					s.Result = "<span style='color:red'>FAILED: " + e.Error() + "</span>"
				}
			}
			conn.Close()
			
			if strings.Contains(string(r), s.Response) {
				s.Result = "<span style='color:green'>OK</span>"
			} else {
				s.Result = "<span style='color:orange'>WARNING: Unexpected response</span>"
			}
		}
	} else {
		conn, err := net.DialTimeout("tcp", s.Host + ":" + strconv.Itoa(s.Port), 10*time.Second)
		if err != nil {
			s.Result = "<span style='color:red'>FAILED: " + err.Error() + "</span>"
		} else {
			if s.Request != "" {
				fmt.Fprintf(conn, s.Request)
			}

			var r = make([]byte, len(s.Response))
			_,e := conn.Read(r)
			if e != nil {
				if e != io.EOF {
					s.Result = "<span style='color:red'>FAILED: " + e.Error() + "</span>"
				}
			}
			conn.Close()

			if strings.Contains(string(r), s.Response) {
				s.Result = "<span style='color:green'>OK</span>"
			} else {
				s.Result = "<span style='color:orange'>WARNING: Unexpected response</span>"
			}
		}
	}

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

	/* Display results in original order */
	var output = "<table>\n"

	var sId = 0
	var hId = 0

	/* Loop hosts */
	for _,h := range conf.Hosts {
		output = output + "<tr><td><strong>" + h.Name + "</strong></td></tr>\n"

		/* Loop services */
		for range h.Services {
			for _,r := range resultSet {
				if r.HostId == hId && r.Id == sId {
					output = output + "<tr><td>" + r.Name + "</td><td>" + r.Result + "</td></tr>\n"
				}
			}
			sId++
		}
		hId++
	}

	output = output + "</table>\n"
	fmt.Fprintf(w, output)
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