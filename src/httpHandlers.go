package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func registerEndpoints(router *mux.Router) *mux.Router {
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", readCoil).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", writeCoil).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/discreteInput/{adr:[0-9]+}", readInput).Methods("GET")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", readHRegister).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", writeHRegister).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/inputRegister/{adr:[0-9]+}", readIRegister).Methods("GET")

	router.HandleFunc("/api/v1/config", getConfig).Methods("GET")
	router.HandleFunc("/api/v1/config", setConfig).Methods("POST")
	return router
}

// FCode 01
func readCoil(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 1
	makeRequest(responseWriter, request, params)
}

// FCode 05/15
func writeCoil(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 5
	if params.Quantity > 1 {
		params.FCode = 15
	}
	makeRequest(responseWriter, request, params)
}

// FCode 02
func readInput(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 2
	makeRequest(responseWriter, request, params)
}

// FCode 03
func readHRegister(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 3
	makeRequest(responseWriter, request, params)
}

// FCode 06/16
func writeHRegister(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 6
	if params.Quantity > 1 {
		params.FCode = 16
	}
	makeRequest(responseWriter, request, params)
}

// FCode 04
func readIRegister(responseWriter http.ResponseWriter, request *http.Request) {
	params := getParams(request)
	params.FCode = 4
	makeRequest(responseWriter, request, params)
}

// Return the current Modbus Config
func getConfig(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("%v %v", request.Method, request.URL)

	data, err := json.Marshal(config)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(data)
}

// Set the current Modbus Config
func setConfig(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("%v %v", request.Method, request.URL)

	var nConf Config
	err := json.NewDecoder(request.Body).Decode(&nConf)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	if nConf.BaudRate != 0 {
		config.BaudRate = nConf.BaudRate
	}
	if nConf.DataBits != 0 {
		config.DataBits = nConf.DataBits
	}
	if nConf.StopBits != 0 {
		config.StopBits = nConf.StopBits
	}
	if nConf.Parity != "" {
		config.Parity = nConf.Parity
	}
	if nConf.Timeout != 0 {
		config.Timeout = nConf.Timeout
	}

	log.Printf("Changed config to (%v Bd, %v parity, %v databits,  %v stopbits, %vms timeout)\n", config.BaudRate, config.Parity, config.DataBits, config.StopBits, config.Timeout)

	data, err := json.Marshal(config)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(data)
}

// getParams extracts all Parameters of a request from the URL
func getParams(request *http.Request) (params Request) {
	{ // Parse register address
		value, err := strconv.ParseUint(mux.Vars(request)["adr"], 10, 16)
		if err != nil {
			println("Err!")
		}
		params.Address = uint16(value)
	}

	{ // Parse serverID
		value, err := strconv.Atoi(mux.Vars(request)["sid"])
		if err != nil {
			println("Err!")
		}
		params.ServerID = byte(value)
	}

	{ //Parse data bytes
		value := request.URL.Query().Get("Data")
		asd, _ := hex.DecodeString(value)
		if len(asd) > 0 {
			params.Data = asd
		}
	}

	{ // Parse 2 value Bytes
		value, _ := strconv.ParseUint(request.URL.Query().Get("Value"), 0, 64)
		params.Value = uint16(value)
	}

	{ // Parse amount of registers affected
		value, err := strconv.Atoi(request.URL.Query().Get("Quantity"))
		if err != nil {
			params.Quantity = 1
		} else {
			params.Quantity = uint16(value)
		}
	}

	// set config for request
	params.Conf.SerialPort = config.SerialPort
	{ // change baudRate
		value, err := strconv.Atoi(request.URL.Query().Get("baudRate"))
		if err != nil {
			params.Conf.BaudRate = config.BaudRate
		} else {
			params.Conf.BaudRate = value
		}
	}
	{ // change dataBits
		value, err := strconv.Atoi(request.URL.Query().Get("dataBits"))
		if err != nil {
			params.Conf.DataBits = config.DataBits
		} else {
			params.Conf.DataBits = value
		}
	}
	{ // change parity
		value := request.URL.Query().Get("parity")
		if len(value) > 0 {
			params.Conf.Parity = value
		} else {
			params.Conf.Parity = config.Parity
		}
	}
	{ // change stopBits
		value, err := strconv.Atoi(request.URL.Query().Get("stopBits"))
		if err != nil {
			params.Conf.StopBits = config.StopBits
		} else {
			params.Conf.StopBits = value
		}
	}
	{ // change timeout
		value, err := strconv.Atoi(request.URL.Query().Get("timeout"))
		if err != nil {
			params.Conf.Timeout = config.Timeout
		} else {
			params.Conf.Timeout = value
		}
	}

	return
}

// make Request forms the callback function and sends the request into the queue
func makeRequest(responseWriter http.ResponseWriter, request *http.Request, params Request) {
	log.Printf("%v %v", request.Method, request.URL)
	var sync = make(chan bool)

	confStr, err := json.Marshal(params.Conf)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("Config", base64.StdEncoding.EncodeToString(confStr))

	// create callback method
	params.Cb = func(res []byte, err error) {
		if err != nil {
			if res != nil { // Serial dev. not found
				responseWriter.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(responseWriter, err)
			} else { // modbus malformed or timeout
				responseWriter.WriteHeader(http.StatusUnprocessableEntity)
				fmt.Fprintln(responseWriter, err)
			}
		} else { // modbus response recieved
			responseWriter.WriteHeader(http.StatusOK)
			fmt.Fprintf(responseWriter, "%s", hex.EncodeToString(res))
		}
		sync <- true // needed to block method
	}
	requestQueue <- params // put request into queue
	<-sync                 // blocks until modbus response was recieved and http response was written
}
