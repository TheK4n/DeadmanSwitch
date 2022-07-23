all: build


build:
	go build -o deadman client.go utils.go
	go build -o deadman-server server.go utils.go

install:
	@if [ "$(shell id -u)" != "0" ]; then echo "Error: Run this target as root" >&2; exit 1; fi
	groupadd -g 1015 deadman
	mkdir /var/lib/deadman-switch
	chmod 700 /var/lib/deadman-switch
	install -m 750 -g deadman -o root ./deadman /usr/bin/deadman
	install -m 700 -o root ./deadman-server /usr/bin/deadman-server

clean:
	@if [ "$(shell id -u)" != "0" ]; then echo "Error: Run this target as root" >&2; exit 1; fi
	rm -f deadman deadman-server

systemd:
	@if [ "$(shell id -u)" != "0" ]; then echo "Error: Run this target as root" >&2; exit 1; fi
	cp ./deadman-switch.service /etc/systemd/system/deadman-switch.service
	systemctl enable deadman-switch
	systemctl start deadman-switch

uninstall:
	@if [ "$(shell id -u)" != "0" ]; then echo "Error: Run this target as root" >&2; exit 1; fi
	systemctl disable deadman-switch
	systemctl stop deadman-switch
	groupdel deadman
	rm /usr/bin/deadman /usr/bin/deadman-server /var/lib/deadman-switch
	rm /etc/systemd/system/deadman-switch.service
