package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var requestQueue = make(chan Request)
var config Config

func usage() {
	fmt.Printf("Usage: %s [optional arguments] serialDevice\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	const (
		defaultPort = 8080
		//defaultSPort    = ""
		defaultBaudRate = 19200
		defaultParity   = "N"
		defaultStopBits = 2
		defaultTimeout  = 1000
		defaultDataBits = 8
		//defaultLocal    = false
	)

	// Modbus Config
	var timeout int
	var port int

	// setup commandline arguments
	flag.Usage = usage
	flag.IntVar(&port, "P", defaultPort, "port")
	flag.IntVar(&config.BaudRate, "b", defaultBaudRate, "serial baud rate ")
	flag.StringVar(&config.Parity, "p", defaultParity, "parity: N - None, E - Even, O - Odd \n(The use of no parity requires 2 stop bits.)")
	flag.IntVar(&config.StopBits, "s", defaultStopBits, "modbus stop bits: 1 or 2")
	flag.IntVar(&config.DataBits, "d", defaultDataBits, "modbus data bits: 5, 6, 7 or 8")
	flag.IntVar(&timeout, "t", defaultTimeout, "serial timeout in ms")
	local := flag.Bool("l", false, "If the flag is set, the server is only available from localhost.")
	flag.Parse()
	config.Timeout = timeout

	// check if parity argument is correct
	if config.Parity != "N" && config.Parity != "E" && config.Parity != "O" {
		flag.Usage()
		os.Exit(1)
	}

	// check if serial port was given
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	config.SerialPort = flag.Args()[0]

	// Http config
	router := registerEndpoints(mux.NewRouter().StrictSlash(true))

	// Start Modbus Routine
	go Run(requestQueue)
	log.Printf("Started Modbus Client (%v, %v Bd, %v parity, %v stopbits, %v Timeout)\n", config.SerialPort, config.BaudRate, config.Parity, config.StopBits, config.Timeout)

	var address string
	if *local {
		address += "localhost"
	}
	address += fmt.Sprintf(":%v", port)
	log.Printf("Started http Server %v\n", address)
	log.Fatal(http.ListenAndServe(address, router))
}
