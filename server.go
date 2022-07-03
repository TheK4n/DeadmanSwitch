package main

import (
    "net"
    // "flag"
    "log"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
    "syscall"
)


func main() {

    socketPath := "/tmp/deadman.socket"


    log.Printf("server starts")
    syscall.Unlink(socketPath)
    listener, _ := net.Listen("unix", socketPath)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        go HandleClient(conn)
    }
}

func hashPassphrase(passphrase string) string {
    h := sha256.New()
    h.Write([]byte(passphrase))
    return hex.EncodeToString(h.Sum(nil))
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

        if hashPassphrase(string(buf[:readLen])) == "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8" {
            conn.Write([]byte("Accepted")) // пишем в сокет
        } else {
            conn.Write([]byte("Declined")) // пишем в сокет
        }
        conn.Close()
        break
    }
}
