all: build


build:
	CGO_ENABLED=0 go build -ldflags=-w -o deadman ./client
	CGO_ENABLED=0 go build -ldflags=-w -o deadmand ./server

clean:
	rm -f deadman deadmand