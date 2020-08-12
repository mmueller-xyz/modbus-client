package modhandler

import (
	"log"
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
}

// The Config struct sets up the serial configuration
type Config struct {
	SerialPort string
	BaudRate   int
	DataBits   int
	Parity     string
	StopBits   int
	Timeout    time.Duration
}

// NewConfig returns a default Config
func NewConfig() (c Config) {
	c.SerialPort = "/dev/ttyUSB0"
	c.BaudRate = 19200
	c.DataBits = 8
	c.Parity = "N"
	c.StopBits = 2
	c.Timeout = 1 * time.Second

	return c
}

// Run starts the modbus client
func Run(rQueue chan Request, conf Config) {
	{ // check if serial device is valid
		h := setupHandler(conf)
		err := h.Connect()

		if err != nil {
			log.Fatal(err)
		}
	}

	for { // main loop
		handleRequest(<-rQueue, conf) // blocks until request is made
	}
}

// setupHandler converts our config to the modbus library's config
func setupHandler(c Config) *modbus.RTUClientHandler {
	h := modbus.NewRTUClientHandler(c.SerialPort)
	h.BaudRate = c.BaudRate
	h.DataBits = c.DataBits
	h.Parity = c.Parity
	h.StopBits = c.StopBits
	h.Timeout = c.Timeout
	// h.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)

	return h
}

// handleRequest is called when a request is made
func handleRequest(r Request, conf Config) {
	var res []byte
	h := setupHandler(conf)
	h.SlaveId = r.ServerID

	err := h.Connect()

	// exit if serial device was not found
	if err != nil {
		res = []byte{0x1} // indicate, that the error is because of the serial dev.
		r.Cb(res, err)
		h.Close()
		return
	}

	c := modbus.NewClient(h)

	switch r.FCode { // call method corresponding to function code
	case 01:
		res, err = c.ReadCoils(r.Address, r.Quantity)
		break
	case 02:
		res, err = c.ReadDiscreteInputs(r.Address, r.Quantity)
		break
	case 03:
		res, err = c.ReadHoldingRegisters(r.Address, r.Quantity)
		break
	case 04:
		res, err = c.ReadInputRegisters(r.Address, r.Quantity)
		break
	case 05:
		res, err = c.WriteSingleCoil(r.Address, r.Value)
		break
	case 06:
		res, err = c.WriteSingleRegister(r.Address, r.Value)
		break
	case 15:
		res, err = c.WriteMultipleCoils(r.Address, r.Quantity, r.Data)
		break
	case 16:
		res, err = c.WriteMultipleRegisters(r.Address, r.Quantity, r.Data)
	}
	h.Close()

	r.Cb(res, err) // call the callback function
}
