package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	modhandler "gitlab.com/enomics/modbus-client/enom-modbus"
)

var ch = make(chan modhandler.Request)

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

func makeRequest(w http.ResponseWriter, r *http.Request, FCode uint16) {
	rd := getParams(r)
	rd.FCode = FCode
	var sync = make(chan bool)

	rd.Cb = func(res []byte, err error) {
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintln(w, err)

		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", hex.EncodeToString(res))
		}
		sync <- true
	}
	ch <- rd
	<-sync
}

func readCoil(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 1)
}
func writeCoil(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 5)
}
func readInput(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 2)
}

func readHRegister(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 3)
}
func writeHRegister(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 6)
}
func readIRegister(w http.ResponseWriter, r *http.Request) {
	makeRequest(w, r, 4)
}

func main() {

	// Modbus Config
	c := modhandler.NewConfig()
	c.BaudRate = 2400
	c.Timeout = 5 * time.Second

	// Http config
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", readCoil).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/coil/{adr:[0-9]+}", writeCoil).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/discreteInput/{adr:[0-9]+}", readInput).Methods("GET")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", readHRegister).Methods("GET")
	router.HandleFunc("/api/v1/{sid:[0-9]+}/holdingRegister/{adr:[0-9]+}", writeHRegister).Methods("POST")

	router.HandleFunc("/api/v1/{sid:[0-9]+}/inputRegister/{adr:[0-9]+}", readIRegister).Methods("GET")

	// Start Concurrent goroutines
	go modhandler.Run(ch, c)

	log.Fatal(http.ListenAndServe(":8080", router))
}
