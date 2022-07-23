package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("unix", SOCKET_FILE)

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	reply := make([]byte, 1024)
	conn.Read(reply)
	fmt.Println(string(reply))

	conn.Write([]byte(secureGetPassword()))

	reply2 := make([]byte, 1024)
	conn.Read(reply2)
	fmt.Println(string(reply2))
}
