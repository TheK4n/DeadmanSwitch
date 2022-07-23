all: build

build:
	go build -o deadman client.go utils.go
	go build -o deadman-server server.go utils.go

install:
	groupadd -g 1015 deadman
	mkdir /var/lib/deadman-switch
	chmod 700 /var/lib/deadman-switch
	install -m 750 -g deadman -o root ./deadman /usr/bin/deadman
	install -m 700 -o root ./deadman-server /usr/bin/deadman-server

clean:
	rm deadman deadman-server

systemd:
	cp ./deadman-switch.service /etc/systemd/system/deadman-switch.service
	systemctl enable deadman-switch
	systemctl start deadman-switch

uninstall:
	rm /usr/bin/deadman /usr/bin/deadman-server /var/lib/deadman-switch
	groupdel deadman
	systemctl disable deadman-switch
	systemctl stop deadman-switch
	rm /etc/systemd/system/deadman-switch.service
