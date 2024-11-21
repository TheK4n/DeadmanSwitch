all: build


build:
	GO111MODULE=off go build -o deadman ./client
	GO111MODULE=off go build -o deadmand ./server

clean:
	rm -f deadman deadman-server