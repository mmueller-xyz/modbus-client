package modhandler

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
func NewConfig() Config {
	var c Config
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
	handler := setupHandler(conf)

	err := handler.Connect()
	defer handler.Close()

	// exit if connection could not be established
	if err != nil {
		panic(err)
	}

	// main loop
	for {
		request := <-rQueue
		handleRequest(request, handler)
	}
}

func setupHandler(c Config) *modbus.RTUClientHandler {
	h := modbus.NewRTUClientHandler(c.SerialPort)
	h.BaudRate = c.BaudRate
	h.DataBits = c.DataBits
	h.Parity = c.Parity
	h.StopBits = c.StopBits
	h.Timeout = c.Timeout
	h.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)

	return h
}

func handleRequest(r Request, h *modbus.RTUClientHandler) {
	h.SlaveId = r.ServerID
	c := modbus.NewClient(h)
	var res []byte
	var err error

	switch r.FCode {
	case 0x01:
		res, err = c.ReadCoils(r.Address, r.Quantity)
		break
	case 0x02:
		res, err = c.ReadDiscreteInputs(r.Address, r.Quantity)
		break
	case 0x03:
		res, err = c.ReadHoldingRegisters(r.Address, r.Quantity)
		break
	case 0x04:
		res, err = c.ReadInputRegisters(r.Address, r.Quantity)
		break
	case 0x05:
		res, err = c.WriteSingleCoil(r.Address, r.Value)
		break
	case 0x06:
		res, err = c.WriteSingleRegister(r.Address, r.Value)
		break
	case 0x15:
		res, err = c.WriteMultipleCoils(r.Address, r.Quantity, r.Data)
		break
	case 0x16:
		res, err = c.WriteMultipleRegisters(r.Address, r.Quantity, r.Data)
	}

	r.Cb(res, err)
}
