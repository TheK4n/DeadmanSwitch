package main

import "fmt"

// var SOCKET_FILE string = "/var/run/deadman.sock"
var SOCKET_FILE string = "/tmp/deadman.sock"

func secureGetPassword() string {
    var input string
    fmt.Print("\033[8m") // Hide input
    fmt.Scanf("%s", &input)
    fmt.Print("\033[28m") // Show input
    return input
}
