package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	modhandler "gitlab.com/enomics/modbus-client/enom-modbus"
)

var rQueue = make(chan modhandler.Request)

// getParams extracts all Parameters of a request from the URL
func getParams(r *http.Request) (rd modhandler.Request) {
	{
		asd, err := strconv.ParseUint(mux.Vars(r)["adr"], 10, 16)
		if err != nil {
			println("Err!")
		}
		rd.Address = uint16(asd)
	}

	{
		asd, err := strconv.Atoi(mux.Vars(r)["sid"])
		if err != nil {
			println("Err!")
		}
		rd.ServerID = byte(asd)
	}

	{
		a := r.URL.Query().Get("Data")
		asd, _ := hex.DecodeString(a)
		if len(asd) > 0 {
			rd.Data = asd
		}
	}
	{
		asd, _ := strconv.Atoi(r.URL.Query().Get("Value"))
		rd.Value = uint16(asd)
	}
	{
		asd, err := strconv.Atoi(r.URL.Query().Get("Quantity"))
		if err != nil {
			rd.Quantity = 1
		} else {
			rd.Quantity = uint16(asd)
		}
	}
	return
}

// make Request forms the callback function and sends the request into the queue
func makeRequest(w http.ResponseWriter, r *http.Request, rd modhandler.Request) {
	log.Println(r.URL)
	var sync = make(chan bool)

	// create callback method
	rd.Cb = func(res []byte, err error) {
		if err != nil {
			if res != nil { // Serial dev. not found
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Serial Device not Found!")
			} else { // modbus malformed or timeout
				w.WriteHeader(http.StatusUnprocessableEntity)
				fmt.Fprintln(w, err)
			}
		} else { // modbus response recieved
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", hex.EncodeToString(res))
		}
		sync <- true // needed to block method
	}
	rQueue <- rd // put request into queue
	<-sync       // blocks until modbus response was recieved and http response was written
}

// FCode 01
func readCoil(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 1
	makeRequest(w, r, rd)
}

// FCode 05
func writeCoil(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 5
	if rd.Quantity > 1 {
		rd.FCode = 15
	}
	makeRequest(w, r, rd)
}

// FCode 02
func readInput(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 2
	makeRequest(w, r, rd)
}

// FCode 03
func readHRegister(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 3
	makeRequest(w, r, rd)
}

// FCode 06
func writeHRegister(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 6
	if rd.Quantity > 1 {
		rd.FCode = 16
	}
	makeRequest(w, r, rd)
}

// FCode 04
func readIRegister(w http.ResponseWriter, r *http.Request) {
	rd := getParams(r)
	rd.FCode = 4
	makeRequest(w, r, rd)
}

func usage() {
	fmt.Printf("Usage: %s [optional arguments] serialDevice\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	const (
		defaultPort     = 8080
		defaultSPort    = ""
		defaultBaudRate = 9600
		defaultParity   = "N"
		defaultStopBits = 2
		defaultTimeout  = 1000
	)

	// Modbus Config
	var c modhandler.Config
	var timeout int
	var port int

	// setup commandline arguments
	flag.Usage = usage
	flag.IntVar(&port, "P", defaultPort, "Port")
	flag.IntVar(&c.BaudRate, "b", defaultBaudRate, "Baud Rate ")
	flag.StringVar(&c.Parity, "p", defaultParity, "Parity: N - None, E - Even, O - Odd \n(The use of no parity requires 2 stop bits.)")
	flag.IntVar(&c.StopBits, "s", defaultStopBits, "Stop bits: 1 or 2")
	flag.IntVar(&timeout, "t", defaultTimeout, "Timeout in ms")
	flag.Parse()
	c.Timeout = time.Duration(timeout) * time.Millisecond

	// check if parity argument is correct
	if c.Parity != "N" && c.Parity != "E" && c.Parity != "O" {
		flag.Usage()
		os.Exit(1)
	}

	// check if serial port was given
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	c.SerialPort = flag.Args()[0]

	// Http config
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", readCoil).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", writeCoil).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/discreteInput/{adr:[0-9]+}", readInput).Methods("GET")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", readHRegister).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", writeHRegister).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/inputRegister/{adr:[0-9]+}", readIRegister).Methods("GET")

	// Start Modbus Routine
	go modhandler.Run(ch, c)
	log.Printf("Started Modbus Client (%v, %v Bd, %v parity, %v stopbits, %v Timeout)\n", c.SerialPort, c.BaudRate, c.Parity, c.StopBits, c.Timeout)

	log.Printf("Started http Server on Port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}
