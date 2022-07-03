all: build

build:
	go build -o deadman client.go
	go build -o deadman-server server.go
