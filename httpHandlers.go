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
	modhandler "gitlab.com/enomics/modbus-client/enom-modbus"
)

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

// Return the current Modbus Config
func getConfig(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL)

	data, err := json.Marshal(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Set the current Modbus Config
func setConfig(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL)

	var nConf modhandler.Config
	err := json.NewDecoder(r.Body).Decode(&nConf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if nConf.BaudRate != 0 {
		c.BaudRate = nConf.BaudRate
	}
	if nConf.DataBits != 0 {
		c.DataBits = nConf.DataBits
	}
	if nConf.StopBits != 0 {
		c.StopBits = nConf.StopBits
	}
	if nConf.Parity != "" {
		c.Parity = nConf.Parity
	}
	if nConf.Timeout != 0 {
		c.Timeout = nConf.Timeout
	}

	log.Printf("Changed config to (%v Bd, %v parity, %v databits,  %v stopbits, %vms timeout)\n", c.BaudRate, c.Parity, c.DataBits, c.StopBits, c.Timeout)

	data, err := json.Marshal(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// getParams extracts all Parameters of a request from the URL
func getParams(r *http.Request) (rd modhandler.Request) {
	{ // Parse register address
		asd, err := strconv.ParseUint(mux.Vars(r)["adr"], 10, 16)
		if err != nil {
			println("Err!")
		}
		rd.Address = uint16(asd)
	}

	{ // Parse serverID
		asd, err := strconv.Atoi(mux.Vars(r)["sid"])
		if err != nil {
			println("Err!")
		}
		rd.ServerID = byte(asd)
	}

	{ //Parse data bytes
		a := r.URL.Query().Get("Data")
		asd, _ := hex.DecodeString(a)
		if len(asd) > 0 {
			rd.Data = asd
		}
	}

	{ // Parse 2 value Bytes
		asd, _ := strconv.Atoi(r.URL.Query().Get("Value"))
		rd.Value = uint16(asd)
	}

	{ // Parse amount of registers affected
		asd, err := strconv.Atoi(r.URL.Query().Get("Quantity"))
		if err != nil {
			rd.Quantity = 1
		} else {
			rd.Quantity = uint16(asd)
		}
	}

	// set config for request
	rd.Conf.SerialPort = c.SerialPort
	{ // change baudRate
		asd, err := strconv.Atoi(r.URL.Query().Get("baudRate"))
		if err != nil {
			rd.Conf.BaudRate = c.BaudRate
		} else {
			rd.Conf.BaudRate = asd
		}
	}
	{ // change dataBits
		asd, err := strconv.Atoi(r.URL.Query().Get("dataBits"))
		if err != nil {
			rd.Conf.DataBits = c.DataBits
		} else {
			rd.Conf.DataBits = asd
		}
	}
	{ // change parity
		a := r.URL.Query().Get("parity")
		if len(a) > 0 {
			rd.Conf.Parity = a
		} else {
			rd.Conf.Parity = c.Parity
		}
	}
	{ // change stopBits
		asd, err := strconv.Atoi(r.URL.Query().Get("stopBits"))
		if err != nil {
			rd.Conf.StopBits = c.StopBits
		} else {
			rd.Conf.StopBits = asd
		}
	}
	{ // change timeout
		asd, err := strconv.Atoi(r.URL.Query().Get("timeout"))
		if err != nil {
			rd.Conf.Timeout = c.Timeout
		} else {
			rd.Conf.Timeout = asd
		}
	}

	return
}

// make Request forms the callback function and sends the request into the queue
func makeRequest(w http.ResponseWriter, r *http.Request, rd modhandler.Request) {
	log.Printf("%v %v", r.Method, r.URL)
	var sync = make(chan bool)

	confStr, err := json.Marshal(rd.Conf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Config", base64.StdEncoding.EncodeToString(confStr))

	// create callback method
	rd.Cb = func(res []byte, err error) {
		if err != nil {
			if res != nil { // Serial dev. not found
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err)
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
