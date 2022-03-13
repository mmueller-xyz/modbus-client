package main

import (
	"log"
	"os"
	"time"

	modbus "github.com/goburrow/modbus"
)

// Request represents a request from http to modbus
//
// Function Code:
//		 1	Read Coil
// 		 2	Read Discrete Input
// 		 3	Read Holding Registers
// 		 4	Read Input Registers
// 		 5	Write Single Coil
// 		 6	Write Single Holding Register
// 		15	Write Multiple Coils
// 		16	Write Multiple Holding Registers
type Request struct {
	ServerID byte                        // ID of the modbus client
	FCode    uint16                      // Function Code
	Address  uint16                      // Address of the Coil/Register
	Data     []byte                      // Used for bulk writing of registers/coils
	Value    uint16                      // Single Value to be written to a coil/register
	Quantity uint16                      // Amount of Coils/Regiters to read from/write to
	Cb       func(res []byte, err error) // Callback
	Conf     Config                      // On-The-Fly config, only for current request
}

// The Config struct sets up the serial configuration
type Config struct {
	SerialPort string `json:"serialPort"` // serialdevice
	BaudRate   int    `json:"baudRate"`   // Baud Rate
	DataBits   int    `json:"dataBits"`   // Data bits: 5, 6, 7 or 8
	Parity     string `json:"parity"`     // Parity: N - None, E - Even, O - Odd
	StopBits   int    `json:"stopBits"`   // Stop bits: 1 or 2
	Timeout    int    `json:"timeout"`    // Timeout in ms
}

// Run starts the modbus client
func Run(rQueue chan Request) {
	for { // main loop
		handleRequest(<-rQueue) // blocks until request is made
	}
}

// setupHandler converts our config to the modbus library's config
func setupHandler(r Request) *modbus.RTUClientHandler {
	clientHandler := modbus.NewRTUClientHandler(r.Conf.SerialPort)
	clientHandler.BaudRate = r.Conf.BaudRate
	clientHandler.DataBits = r.Conf.DataBits
	clientHandler.Parity = r.Conf.Parity
	clientHandler.StopBits = r.Conf.StopBits
	clientHandler.Timeout = time.Duration(r.Conf.Timeout) * time.Millisecond
	clientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)

	return clientHandler
}

// handleRequest is called when a request is made
func handleRequest(request Request) {
	var response []byte
	clientHandler := setupHandler(request)
	clientHandler.SlaveId = request.ServerID

	err := clientHandler.Connect()

	// exit if serial device was not found
	if err != nil {
		response = []byte{0x1} // indicate, that the error is because of the serial dev.
		request.Cb(response, err)
		clientHandler.Close()
		return
	}

	client := modbus.NewClient(clientHandler)

	switch request.FCode { // call method corresponding to function code
	case 01:
		response, err = client.ReadCoils(request.Address, request.Quantity)
		break
	case 02:
		response, err = client.ReadDiscreteInputs(request.Address, request.Quantity)
		break
	case 03:
		response, err = client.ReadHoldingRegisters(request.Address, request.Quantity)
		break
	case 04:
		response, err = client.ReadInputRegisters(request.Address, request.Quantity)
		break
	case 05:
		response, err = client.WriteSingleCoil(request.Address, request.Value)
		break
	case 06:
		response, err = client.WriteSingleRegister(request.Address, request.Value)
		break
	case 15:
		response, err = client.WriteMultipleCoils(request.Address, request.Quantity, request.Data)
		break
	case 16:
		response, err = client.WriteMultipleRegisters(request.Address, request.Quantity, request.Data)
	}
	clientHandler.Close()

	request.Cb(response, err) // call the callback function
}
