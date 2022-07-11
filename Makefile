all: build

build:
	go build -o deadman client.go utils.go
	go build -o deadman-server server.go utils.go

install:
	cp deadman /usr/bin/deadman
	cp deadman-server /usr/bin/deadman-server
	chmod g+x /usr/bin/deadman /usr/bin/deadman-server
	
	groupadd deadman
	umask 006
	mkdir /var/lib/deadman-switch
	chgrp deadman /var/lib/deadman-switch /usr/bin/deadman /usr/bin/deadman-server

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
