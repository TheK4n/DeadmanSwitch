package main

import (
    "net"
    // "flag"
    "log"
    "fmt"
)


func main() {

    log.Printf("server starts")
    listener, _ := net.Listen("unix", "/tmp/deadman.socket")

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        go HandleClient(conn)
    }
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
        } else {
            conn.Write([]byte("Declined")) // пишем в сокет
        }
        conn.Close()
        break
    }
}
