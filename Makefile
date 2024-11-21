all: build


build:
	CGO_ENABLED=0 GO111MODULE=off go build -ldflags=-w -o deadman ./client
	CGO_ENABLED=0 GO111MODULE=off go build -ldflags=-w -o deadmand ./server

clean:
	rm -f deadman deadmand