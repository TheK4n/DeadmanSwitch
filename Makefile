all: build


build:
	CGO_ENABLED=0 go build -o deadman ./client
	CGO_ENABLED=0 go build -o deadmand ./server

clean:
	rm -f deadman deadmand