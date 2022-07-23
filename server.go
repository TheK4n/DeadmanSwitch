package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/lu4p/shred"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
    "net/http"
    "net/url"
    "encoding/binary"
)

const ONE_MONTH_SEC int = 60 * 60 * 24 * 30
const PREFIX string = "/var/lib/deadman-switch"
const PUBLIC_DIR = PREFIX + "/public"
const PRIVATE_DIR = PREFIX + "/private"
const TIME_FILE = PREFIX + "/time"
const HASH_FILE = PREFIX + "/hash"

var COMMANDS = []string{"run", "init"}

func kill(msg string, code int) {
	err := fmt.Errorf(msg)
	fmt.Println(err.Error())
	os.Exit(code)
}

func parseCommand() string {

	if len(os.Args) < 2 {
		kill("No command", 1)
	}
	command := os.Args[1]
	if !isValidCommand(COMMANDS, command) {
		kill("Wrong command", 1)
	}
	return command
}

// checks is used command are valid
func isValidCommand(commands []string, command string) bool {
	for _, com := range commands {
		if command == com {
			return true
		}
	}
	return false
}
func writeHash(hash string) error {
	return ioutil.WriteFile(HASH_FILE, []byte(hash), 0600)
}

func PowInts(x, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	y := PowInts(x, n/2)
	if n%2 == 0 {
		return y * y
	}
	return x * y * y
}

func hashPassphrase(passphrase string, salt string) string {
	prevHash := passphrase
	iterations := PowInts(2, 16)

	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write([]byte(prevHash + salt))
		prevHash = hex.EncodeToString(h.Sum(nil))
	}
	return prevHash + salt
}

func genTrulyRandom() int64 {
    file, _ := os.Open("/dev/urandom")
    defer file.Close()

    const maxSz = 256
    // create buffer
    b := make([]byte, maxSz)

    // read content to buffer
    file.Read(b)

    return int64(binary.BigEndian.Uint64(b))
}

func generateSalt() string {
	rand.Seed(genTrulyRandom())
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%f", rand.Float64())))
	return hex.EncodeToString(h.Sum(nil))
}

func checkHash(passphrase string) (bool, error) {

	storedHashAndSalt, err := ioutil.ReadFile(HASH_FILE)
	storedSalt := storedHashAndSalt[64:]
	hash := hashPassphrase(passphrase, string(storedSalt))
	return hash == string(storedHashAndSalt), err
}

// asks master passphrase, write hash and updates time
func initialSetup() {
	fmt.Print("Input passphrase: ")
	inputPassphrase := secureGetPassword()
	fmt.Print("Repeat passphrase: ")
	repeatedPassphrase := secureGetPassword()

	if inputPassphrase == repeatedPassphrase {
		err := writeHash(hashPassphrase(inputPassphrase, generateSalt()))
		if err != nil {
			log.Printf("error: writing hash")
		}
		updateTime(ONE_MONTH_SEC)
	} else {
		kill("Passphrases didnt match", 1)
	}
}

func getRestOfTime() int {
	restOfTime, _ := ioutil.ReadFile(TIME_FILE)
	i, _ := strconv.Atoi(string(restOfTime))
	return i
}

func updateTime(seconds int) {
	now := time.Now()
	err := ioutil.WriteFile(TIME_FILE, []byte(fmt.Sprintf("%d", int(now.Unix())+seconds)), 0600)
	if err != nil {
		return
	}
	log.Print("Extended until: " + getCurTime())
}

func initDeadmanSwitch() {
	log.Printf("Deadman Switch EXECUTED!")
	err := shredPrivateFiles()
	if err != nil {
		log.Printf("error: while shreading files")
	}
	err = publicatePublicFiles()
	if err != nil {
		log.Printf("error: while publicating files")
	}
	os.Exit(0)
}

func shredPrivateFiles() error {
	files, err := ioutil.ReadDir(PRIVATE_DIR)
	shredconf := shred.Conf{Times: 2, Zeros: true, Remove: true}

	for _, file := range files {
		shredconf.Path(PRIVATE_DIR + "/" + file.Name())
	}
	return err
}

func publicatePublicFiles() error {
	files, err := ioutil.ReadDir(PUBLIC_DIR)
	fmt.Print("Files to publicate: ")

	for _, file := range files {
        text, _ := ioutil.ReadFile(PUBLIC_DIR + "/" + file.Name())
        sendTelegramMessage(string(text))
	}
	return err
}

func sendTelegramMessage(text string) {

    groupId := os.Getenv("GROUP_ID")
    token := os.Getenv("TOKEN")
    sendMessageUrl := "https://api.telegram.org/bot"+ token + "/sendMessage"

    data := url.Values{
        "chat_id": {groupId},
        "text": {text},
    }

    http.PostForm(sendMessageUrl, data)
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

		isValidHash, err := checkHash(string(buf[:readLen]))
		if isValidHash {
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
        os.Open(SOCKET_FILE)
        os.Chown(SOCKET_FILE, 0, 1015)
        os.Chmod(SOCKET_FILE, 0660)
		log.Printf("Server starts")
		log.Printf("Expires at: " + getCurTime())

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
	syscall.Unlink(SOCKET_FILE) // clean unix socket
	os.Remove(SOCKET_FILE)
}
