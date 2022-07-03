package main

import "fmt"

func secureGetPassword() string {
    var input string
    fmt.Print("\033[8m") // Hide input
    fmt.Scanf("%s", &input)
    fmt.Print("\033[28m") // Show input
    return input
}
