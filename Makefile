all: build

build:
	go build -o deadman client.go utils.go
	go build -o deadman-server server.go utils.go

systemd:
	cp ./deadman-switch.service /etc/systemd/system/deadman-switch.service
	systemd enable deadman-switch
	systemd start deadman-switch
