package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const HELP_MESSAGE string = `deadman - DeadmanSwitch client

Usage: deadmand <COMMAND>

Commands:
	extend - extend server fault timeout
	execute - forcely execute server fault command
`

func main() {
	if len(os.Args) < 2 {
		die(1, "Usage: %s <COMMAND>", os.Args[0])
	}

	socketFile := os.Getenv("DEADMAN_SOCKET")
	if socketFile == "" {
		socketFile = "/tmp/deadman.sock"
	}

	switch os.Args[1] {
	case "execute", "extend":
		conn, err := net.Dial("unix", socketFile)
		if err != nil {
			die(1, "%s", err.Error())
		}

		fmt.Print("Input passphrase: ")
		passphrase := secureGetPassword()
		client := DeadmanClient{conn}
		err = client.sendToServer(os.Args[1], passphrase)
		if err != nil {
			die(1, "%s", err.Error())
		}

	case "--help":
		fmt.Print(HELP_MESSAGE)
		os.Exit(0)
	default:
		die(1, "'%s' is not a %s command", os.Args[1], os.Args[0])
	}
}

func secureGetPassword() string {
	var input string
	fmt.Print("\033[8m")
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m")
	return input
}

type DeadmanClient struct {
	conn net.Conn
}

func (cl *DeadmanClient) sendToServer(command string, passphrase string) error {
	reply := make([]byte, 1024)

	_, err := cl.conn.Read(reply)
	if err != nil {
		return err
	}

	fmt.Println(string(reply))

	messageToServer := command + " " + passphrase
	_, err = cl.conn.Write([]byte(messageToServer))
	if err != nil {
		return err
	}

	reply2 := make([]byte, 1024)
	_, err = cl.conn.Read(reply2)
	if err != nil {
		return err
	}

	fmt.Println(string(reply2))
	return nil
}


func die(code int, format string, a ...any) {
	_, programName := filepath.Split(os.Args[0])

	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s: error: ", programName) + format, a...)
	os.Exit(code)
}