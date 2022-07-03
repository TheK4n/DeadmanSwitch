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
    "time"
    "strconv"
)


var PEPPER string = "cd031f46f24d5bd3543"
var ONE_MONTH_SEC int = 2592000
var TIME_FILE string = ".time.txt"
var HASH_FILE string = ".hash.txt"
var SOCKET_FILE string = "/tmp/deadman.socket"


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
    ioutil.WriteFile(HASH_FILE, []byte(hash), 0644)
}

func hashPassphrase(passphrase string, salt string) string {
    h := sha256.New()
    h.Write([]byte(passphrase + salt + PEPPER))
    return hex.EncodeToString(h.Sum(nil)) + salt
}

func generateSalt() string {
    now := time.Now()
    nanoSec := now.UnixNano()
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%d", nanoSec)))
    return hex.EncodeToString(h.Sum(nil))
}

func checkHash(passphrase string) bool {

    storedHashAndSalt, _ := ioutil.ReadFile(HASH_FILE)

    storedSalt := storedHashAndSalt[64:]

    hash := hashPassphrase(passphrase, string(storedSalt))

    return hash == string(storedHashAndSalt)
}

func initialSetup() {
    fmt.Print("Input passphrase: ")
    inputPassphrase := secureGetPassword()

    fmt.Println("")

    fmt.Print("Repeat passphrase: ")
    repeatedPassphrase := secureGetPassword()

    if inputPassphrase == repeatedPassphrase {
        writeHash(hashPassphrase(inputPassphrase, generateSalt()))
        updateTime(ONE_MONTH_SEC)
    } else {
        panic("Passphrases didnt match")
    }


}

func getRestOfTime() int {
    restOfTime, _ := ioutil.ReadFile(TIME_FILE)
    i, _ := strconv.Atoi(string(restOfTime))
    return i
}

func updateTime(seconds int) {
    now := time.Now()
    ioutil.WriteFile(TIME_FILE, []byte(fmt.Sprintf("%d", int(now.Unix()) + seconds)), 0644)
}

func initDeadmanSwitch() {
    log.Printf("Deadman Switch EXECUTED!")
    os.Exit(0)
}

func timeout() {
    for {
        time.Sleep(15 * time.Second)

        now := time.Now()

        if getRestOfTime() < int(now.Unix()) {
            initDeadmanSwitch()
            break
        }

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

        if checkHash(string(buf[:readLen])) {
            updateTime(ONE_MONTH_SEC)
            conn.Write([]byte("Extended until: " + fmt.Sprintf("%s", time.Unix(int64(getRestOfTime()), 0))))
        } else {
            conn.Write([]byte("Declined, expires at: " + fmt.Sprintf("%s", time.Unix(int64(getRestOfTime()), 0))))
        }
        conn.Close()
        break
    }
}

func main() {

    command := parseCommand()

    switch command {
        case "run":
            go timeout()
            syscall.Unlink(SOCKET_FILE) // clean unix socket

            listener, _ := net.Listen("unix", SOCKET_FILE)
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

