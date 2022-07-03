package main

import (
    "net"
    "log"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
    "syscall"
    "io/ioutil"
    "os"
)


var PEPPER string = "cd031f46f24d5bd3543"

func isValidCommand(commands []string, command string) bool {
    for _, com := range commands {
        if command == com {
            return true
        }
    }
    return false
}

func parseCommand() string {

    if len(os.Args) < 2 {
        panic("Wrong command")
    }

    command := os.Args[1]

    if !isValidCommand([]string{"run", "init"}, command) {
        panic("Wrong command")
    }
    return command

}

func writeHash(hash string) {
    ioutil.WriteFile("hash.txt", []byte(hash), 0644)
}

func hashPassphrase(passphrase string) string {
    h := sha256.New()
    h.Write([]byte(passphrase + PEPPER))
    return hex.EncodeToString(h.Sum(nil))
}

func initialSetup() {
    fmt.Print("Input passphrase: ")
    inputPassphrase := secureGetPassword()

    fmt.Println("")

    fmt.Print("Repeat passphrase: ")
    repeatedPassphrase := secureGetPassword()

    if inputPassphrase == repeatedPassphrase {
        writeHash(hashPassphrase(inputPassphrase))
    } else {
        panic("Passphrases didnt match")
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

        hash := hashPassphrase(string(buf[:readLen]))

        storedHash, _ := ioutil.ReadFile("hash.txt")

        if hash == string(storedHash) {
            conn.Write([]byte("Accepted")) // пишем в сокет
        } else {
            conn.Write([]byte("Declined")) // пишем в сокет
        }
        conn.Close()
        break
    }
}

func main() {
    command := parseCommand()

    switch command {
        case "run":
            socketPath := "/tmp/deadman.socket"
            syscall.Unlink(socketPath) // clean unix socket

            listener, _ := net.Listen("unix", socketPath)
            log.Printf("Server starts")

            for {
                conn, err := listener.Accept()
                if err != nil {
                    continue
                }

                go HandleClient(conn)
            }
        case "init":
            initialSetup()
    }
}

