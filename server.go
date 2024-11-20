package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
    "os/exec"
	"strconv"
	"syscall"
	"time"
)

const ONE_MONTH_SEC int = 60 //60 * 60 * 24 * 30
const _PREFIX string = "/home/thek4n/.local/deadman"
const PUBLIC_DIR = _PREFIX + "/public"
const PRIVATE_DIR = _PREFIX + "/private"
const TIME_FILE = _PREFIX + "/time"
const HASH_FILE = _PREFIX + "/hash"
var COMMANDS = []string{"run", "init"}

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
	return os.WriteFile(HASH_FILE, []byte(hash), 0600)
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
    r := rand.New(rand.NewSource(genTrulyRandom()))
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%f", r.Float64())))
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
	storedHashAndSalt, err := os.ReadFile(HASH_FILE)
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
	restOfTime, _ := os.ReadFile(TIME_FILE)
	i, _ := strconv.Atoi(string(restOfTime))
	return i
}

func updateTime(seconds int) {
	now := time.Now()
	err := os.WriteFile(TIME_FILE, []byte(fmt.Sprintf("%d", int(now.Unix())+seconds)), 0600)
	if err != nil {
		return
	}
	log.Print("Extended until: " + getCurTime())
}

func initDeadmanSwitch() {
	log.Printf("Deadman Switch EXECUTED!")

    cmd := exec.Command("touch", "/home/thek4n/DEADMAN")
    if err := cmd.Run(); err != nil {
        fmt.Println("Error: ", err)
    }

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
