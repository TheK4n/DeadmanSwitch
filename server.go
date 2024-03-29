package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/lu4p/shred" // unix commmand shred
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"
)

const ONE_MONTH_SEC int = 60 * 60 * 24 * 30
const _PREFIX string = "/var/lib/deadman-switch"
const PUBLIC_DIR = _PREFIX + "/public"
const PRIVATE_DIR = _PREFIX + "/private"
const TIME_FILE = _PREFIX + "/time"
const HASH_FILE = _PREFIX + "/hash"
const COMMANDS = []string{"run", "init"}

func parseCommand() string {

	if len(os.Args) < 2 {
		kill("Usage: deadman-server COMMAND", 1)
	}
	command := os.Args[1]
	if !isValidCommand(COMMANDS, command) {
		kill("'"+os.Args[1]+"' is not a deadman-server command.", 1)
	}
	return command
}

// check is used command are valid
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

func hashPassphrase(passphrase, salt string) string {
	prevHash := passphrase
	iterations := PowInts(2, 16)

	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write([]byte(prevHash + salt))
		prevHash = hex.EncodeToString(h.Sum(nil))
	}
	return prevHash + salt
}

func generateSalt() string {
	rand.Seed(genTrulyRandom())
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%f", rand.Float64())))
	return hex.EncodeToString(h.Sum(nil))
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
	sendMessageUrl := "https://api.telegram.org/bot" + token + "/sendMessage"

	data := url.Values{
		"chat_id": {groupId},
		"text":    {text},
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

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 32) // buffer for client data
	for {
		conn.Write([]byte("Write passphrase:"))

		readLen, err := conn.Read(buf)
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

			go handleClient(conn)
		}
	case "init":
		initialSetup()
	}
	syscall.Unlink(SOCKET_FILE) // clean unix socket
	os.Remove(SOCKET_FILE)
}
