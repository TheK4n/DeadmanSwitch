package common

import (
	"fmt"
	"os"
)

const DEADMAN_SOCKET_VARIABLE_NAME = "DEADMAN_SOCKET"

func GetSocketPath() string {
	socketEnv := os.Getenv(DEADMAN_SOCKET_VARIABLE_NAME)

	if socketEnv != "" {
		return socketEnv
	}
	return "/tmp/deadman.sock"
}

func SecureGetPassword() string {
	var input string
	fmt.Print("\033[8m")
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m")
	return input
}

func Die(msg string, code int) {
	err := fmt.Errorf(msg)
	fmt.Println("deadman:", err.Error())
	os.Exit(code)
}