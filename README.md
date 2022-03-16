# Modbus Client
This Modbus Client serves as a Middleware, translating HTTP requests to ModbusRTU serial requests. In order to maximize serial usage, a queue is used to keep the serial worker always occupied.

# Dependencies
```shell script
go get github.com/goburrow/modbus
go get github.com/gorilla/mux
```

# Usage
```shell script
Usage: ./modbus-client [optional arguments] serialDevice
  -P int
        serial port (default 8080)
  -b int
        serial baud rate  (default 19200)
  -d int
        modbus data bits: 5, 6, 7 or 8 (default 8)
  -l    If the flag is set, the server is only avalilable from localhost.
  -p string
        parity: N - None, E - Even, O - Odd 
        (The use of no parity requires 2 stop bits.) (default "N")
  -s int
        modbus stop bits: 1 or 2 (default 2)
  -t int
        serial timeout in ms (default 1000)
```

## HTTP endpoints
The endpoints are described in the following yaml file:

[openapi.yaml](./openapi.yaml)

### root
`/api/v1`
