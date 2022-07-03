package main

import (
    "net"
    "flag"
    "log"
    "fmt"
)


func main() {


    host, port := parseParams()
    checkParams(host, port)

    log.Printf("server starts on %s:%d", host, port)
    listener, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        go HandleClient(conn)
    }
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

func HandleClient(conn net.Conn) {
    defer conn.Close()


    buf := make([]byte, 32) // буфер для чтения клиентских данных
    for {
        conn.Write([]byte("Write passphrase:")) // пишем в сокет

        readLen, err := conn.Read(buf) // читаем из сокета
        if err != nil {
            fmt.Println(err)
            break
        }

        if (string(buf[:readLen]) == "password") {
            conn.Write([]byte("Accepted")) // пишем в сокет
            conn.Close()
            break
        } else {
            conn.Write([]byte("Declined")) // пишем в сокет
            conn.Close()
            break
        }
    }
}
