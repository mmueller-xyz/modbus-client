package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	modhandler "gitlab.com/enomics/modbus-client/enom-modbus"
)

var rQueue = make(chan modhandler.Request)
var c modhandler.Config

func usage() {
	fmt.Printf("Usage: %s [optional arguments] serialDevice\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	const (
		defaultPort     = 8080
		defaultSPort    = ""
		defaultBaudRate = 19200
		defaultParity   = "N"
		defaultStopBits = 2
		defaultTimeout  = 1000
		defaultDataBits = 8
		defaultLocal    = false
	)

	// Modbus Config
	var timeout int
	var port int

	// setup commandline arguments
	flag.Usage = usage
	flag.IntVar(&port, "P", defaultPort, "Port")
	flag.IntVar(&c.BaudRate, "b", defaultBaudRate, "Baud Rate ")
	flag.StringVar(&c.Parity, "p", defaultParity, "Parity: N - None, E - Even, O - Odd \n(The use of no parity requires 2 stop bits.)")
	flag.IntVar(&c.StopBits, "s", defaultStopBits, "Stop bits: 1 or 2")
	flag.IntVar(&c.DataBits, "d", defaultDataBits, "Data bits: 5, 6, 7 or 8")
	flag.IntVar(&timeout, "t", defaultTimeout, "Timeout in ms")
	local := flag.Bool("l", false, "If the flag is set, the server is only avalilable from localhost.")
	flag.Parse()
	c.Timeout = timeout

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

	router.HandleFunc("/api/v1/config", getConfig).Methods("GET")
	router.HandleFunc("/api/v1/config", setConfig).Methods("POST")

	// Start Modbus Routine
	go modhandler.Run(rQueue)
	log.Printf("Started Modbus Client (%v, %v Bd, %v parity, %v stopbits, %v Timeout)\n", c.SerialPort, c.BaudRate, c.Parity, c.StopBits, c.Timeout)

	var address string
	if *local {
		address += "localhost"
	}
	address += fmt.Sprintf(":%v", port)
	log.Printf("Started http Server %v\n", address)
	log.Fatal(http.ListenAndServe(address, router))
}
