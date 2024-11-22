package main

import (
	"fmt"
	"net"
	"os"

	common "github.com/thek4n/DeadmanSwitch/common"
)

var SOCKET_FILE = common.GetSocketPath()

func main() {
	handleCommand(parseCommand(os.Args))
}

func parseCommand(args []string) string {
	if len(args) < 2 {
		common.Die("Usage: "+args[0]+" COMMAND", 1)
	}
	command := args[1]
	return command
}

func handleCommand(command string) {
	switch command {
	case "execute":
		sendToServer(command)
	case "extend":
		sendToServer(command)
	case "--help":
		fmt.Println("execute\nextend")
		os.Exit(0)
	default:
		common.Die("'"+os.Args[1]+"' is not a "+os.Args[0]+" command.", 1)
	}
}

func sendToServer(command string) {
	conn, err := net.Dial("unix", SOCKET_FILE)

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	reply := make([]byte, 1024)
	conn.Read(reply)
	fmt.Println(string(reply))

	messageToServer := command + " " + common.SecureGetPassword()
	conn.Write([]byte(messageToServer))

	reply2 := make([]byte, 1024)
	conn.Read(reply2)
	fmt.Println(string(reply2))
}
