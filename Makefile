all: build

build:
	go build -o deadman client.go utils.go
	go build -o deadman-server server.go utils.go
