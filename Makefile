
.phony: build

build: bin/modbus-client bin/modbus-client-windows-amd64 bin/modbus-client-linux-arm64 bin/modbus-client-linux-386 bin/modbus-client-linux-armv7

bin/modbus-client: src/main.go
	cd src; go build -o ../bin/modbus-client
	
bin/modbus-client-windows-amd64: src/main.go
	cd src; GOOS=windows GOARCH=amd64 go build -o ../bin/modbus-client-windows-amd64.exe

bin/modbus-client-linux-386: src/main.go
	cd src; GOOS=linux GOARCH=386 go build -o ../bin/modbus-client-linux-386

bin/modbus-client-linux-arm64: src/main.go
	cd src; GOOS=linux GOARCH=arm64 go build -o ../bin/modbus-client-linux-arm64

bin/modbus-client-linux-armv7: src/main.go
	cd src; GOOS=linux GOARCH=arm GOARM=7 go build -o ../bin/modbus-client-linux-armv7

run:
	go run src/
