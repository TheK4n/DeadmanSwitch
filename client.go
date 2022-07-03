package main

import (
    "net"
    // "flag"
    "fmt"
)

func secureGetPassword() string {
    var input string
    fmt.Print("\033[8m") // Hide input
    fmt.Scanf("%s", &input)
    fmt.Print("\033[28m") // Show input
    return input
}

func main() {
    conn, err := net.Dial("unix", "/tmp/deadman.socket")

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

