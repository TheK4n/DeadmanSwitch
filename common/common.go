package common

import (
	"fmt"
	"os"
)

const SOCKET_FILE string = "/tmp/deadman.sock"

func SecureGetPassword() string {
	var input string
	fmt.Print("\033[8m")
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m")
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

func Kill(msg string, code int) {
	err := fmt.Errorf(msg)
	fmt.Println("deadman:", err.Error())
	os.Exit(code)
}
