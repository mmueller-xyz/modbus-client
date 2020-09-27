package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var rQueue = make(chan Request)
var c Config

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
	flag.IntVar(&c.BaudRate, "b", defaultBaudRate, "serial baud rate ")
	flag.StringVar(&c.Parity, "p", defaultParity, "parity: N - None, E - Even, O - Odd \n(The use of no parity requires 2 stop bits.)")
	flag.IntVar(&c.StopBits, "s", defaultStopBits, "modbus stop bits: 1 or 2")
	flag.IntVar(&c.DataBits, "d", defaultDataBits, "modbus data bits: 5, 6, 7 or 8")
	flag.IntVar(&timeout, "t", defaultTimeout, "serial timeout in ms")
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
	router := registerEndpoints(mux.NewRouter().StrictSlash(true))

	// Start Modbus Routine
	go Run(rQueue)
	log.Printf("Started Modbus Client (%v, %v Bd, %v parity, %v stopbits, %v Timeout)\n", c.SerialPort, c.BaudRate, c.Parity, c.StopBits, c.Timeout)

	var address string
	if *local {
		address += "localhost"
	}
	address += fmt.Sprintf(":%v", port)
	log.Printf("Started http Server %v\n", address)
	log.Fatal(http.ListenAndServe(address, router))
}
