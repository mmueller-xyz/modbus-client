
.phony: build

build: bin/modbus-client bin/modbus-client-linux-arm64 bin/modbus-client-linux-386 bin/modbus-client-linux-armv7

bin/modbus-client: main.go enom-modbus/mod-handler.go
	go build -o bin/modbus-client
	
bin/modbus-client-linux-386: main.go enom-modbus/mod-handler.go
	GOOS=linux GOARCH=386 go build -o bin/modbus-client-linux-386

bin/modbus-client-linux-arm64: main.go enom-modbus/mod-handler.go
	GOOS=linux GOARCH=arm64 go build -o bin/modbus-client-linux-arm64

bin/modbus-client-linux-armv7: main.go enom-modbus/mod-handler.go
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/modbus-client-linux-armv7

run:
	go run main.go
