package main

import (
    "net"
    "flag"
    "fmt"
)


func main() {
    host, port := parseParams()
    checkParams(host, port)


    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))

    if err != nil {
       fmt.Println("error:", err)
       return
    }

    reply := make([]byte, 1024)
    conn.Read(reply)
    fmt.Println(string(reply))

    var input string
    fmt.Scanf("%s", &input)

    conn.Write([]byte(input))

    reply2 := make([]byte, 1024)
    conn.Read(reply2)
    fmt.Println(string(reply2))
}

func checkParams(host string, port int) {
    if !(port > 0 && port < 65535) {
        panic("Wrong port!")
    }
}

func parseParams() (string, int) {
    var host = flag.String("h", "127.0.0.1", "Server host")
    var port = flag.Int("p", 8081, "Server port")

    flag.Parse()
    return *host, *port
}
