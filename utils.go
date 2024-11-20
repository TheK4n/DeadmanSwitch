package main

import (
    "fmt"
    "os"
)

const SOCKET_FILE string = "/tmp/deadman.sock"

// hidden get password from stdin
func secureGetPassword() string {
	var input string
	fmt.Print("\033[8m") // Hide input
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m") // Show input
	return input
}

func PowInts(x, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	y := PowInts(x, n/2)
	if n%2 == 0 {
		return y * y
	}
	return x * y * y
}

// kill program with error
func kill(msg string, code int) {
	err := fmt.Errorf(msg)
	fmt.Println("deadman:", err.Error())
	os.Exit(code)
}
