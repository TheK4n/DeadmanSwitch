package main

import "fmt"

const SOCKET_FILE string = "/var/run/deadman.sock"

func secureGetPassword() string {
	var input string
	fmt.Print("\033[8m") // Hide input
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m") // Show input
	return input
}
