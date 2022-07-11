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
    "math/rand"
    "github.com/lu4p/shred"
)


const ONE_MONTH_SEC int = 60//*60*24*30
const PREFIX string = "/var/lib/deadman-switch"
const PUBLIC_DIR string = PREFIX + "/public"
const PRIVATE_DIR string = PREFIX + "/private"
const TIME_FILE string = PREFIX + "/time"
const HASH_FILE string = PREFIX + "/hash"


// checks is used command are valid
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
    ioutil.WriteFile(HASH_FILE, []byte(hash), 0600)
}

func PowInts(x, n int) int {
   if n == 0 { return 1 }
   if n == 1 { return x }
   y := PowInts(x, n/2)
   if n % 2 == 0 { return y*y }
   return x*y*y
}

func hashPassphrase(passphrase string, salt string) string {
    prevHash := passphrase
    iterations := PowInts(2, 16)

    for i:=0; i < iterations; i++ {
        h := sha256.New()
        h.Write([]byte(prevHash + salt))
        prevHash = hex.EncodeToString(h.Sum(nil))
    }
    return prevHash + salt
}

func generateSalt() string {
    rand.Seed(time.Now().UnixNano())
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%d", rand.Float64)))
    return hex.EncodeToString(h.Sum(nil))
}

func checkHash(passphrase string) bool {

    storedHashAndSalt, _ := ioutil.ReadFile(HASH_FILE) // err!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!11

    storedSalt := storedHashAndSalt[64:]

    hash := hashPassphrase(passphrase, string(storedSalt))

    return hash == string(storedHashAndSalt)
}

// asks master passphrase, write hash and updates time
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
    ioutil.WriteFile(TIME_FILE, []byte(fmt.Sprintf("%d", int(now.Unix()) + seconds)), 0600)
    log.Print("Extended until: " + getCurTime())
}

func initDeadmanSwitch() {
    log.Printf("Deadman Switch EXECUTED!")
    shredPrivateFiles()
    publicatePublicFiles()
    os.Exit(0)
}

func shredPrivateFiles() {
    files, _ := ioutil.ReadDir(PRIVATE_DIR) // err !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
    shredconf := shred.Conf{Times: 2, Zeros: true, Remove: true}

    for _, file := range files {
        shredconf.Path(PRIVATE_DIR + "/" + file.Name())
    }
}

func publicatePublicFiles() {
    files, _ := ioutil.ReadDir(PUBLIC_DIR) // err !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
    fmt.Print("Files to publicate: ")

    for _, file := range files {
        fmt.Print(file.Name() + " ")
    }
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

func getCurTime() string {
    return fmt.Sprintf("%s", time.Unix(int64(getRestOfTime()), 0))
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
            conn.Write([]byte("Extended until: " + getCurTime()))
        } else {
            conn.Write([]byte("Declined, expires at: " + getCurTime()))
        }
        conn.Close()
        break
    }
}

func main() {

    os.Remove(SOCKET_FILE)
    syscall.Unlink(SOCKET_FILE) // clean unix socket
    command := parseCommand()

    switch command {
        case "run":
            go timeout()

            listener, _ := net.Listen("unix", SOCKET_FILE)
            log.Printf("Server starts")
            log.Printf("Expires at: " + getCurTime())

            for {
                conn, err := listener.Accept()
                if err != nil { continue }

                go HandleClient(conn)
            }
        case "init":
            initialSetup()
    }
    syscall.Unlink(SOCKET_FILE) // clean unix socket
    os.Remove(SOCKET_FILE)
}

