package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"path/filepath"

	common "../common"
)

const TIMEOUT_SEC int = 60

var PREFIX = os.Getenv("HOME") + "/.local/deadman"
var TIME_FILE = PREFIX + "/time"
var HASH_FILE = PREFIX + "/hash"


func writeHash(hash string) error {
    hashfile_dir := filepath.Dir(HASH_FILE)
    err := os.MkdirAll(hashfile_dir, 0700)

    if err != nil {
        return err
    }
	return os.WriteFile(HASH_FILE, []byte(hash), 0600)
}

func parseCommand() string {
	if len(os.Args) < 2 {
		common.Kill("Usage: "+os.Args[0]+" COMMAND", 1)
	}
	command := os.Args[1]
	if !isValidCommand([]string{"run", "init"}, command) {
		common.Kill("'"+os.Args[1]+"' is not a "+os.Args[0]+" command.", 1)
	}
	return command
}

func isValidCommand(commands []string, command string) bool {
	for _, com := range commands {
		if command == com {
			return true
		}
	}
	return false
}


func initialSetup() {
    firstPassphrase := askPassphrase()
    secondPassphrase := askPassphrase()

    if ! isPassphrasesMatch(firstPassphrase, secondPassphrase) {
		common.Kill("Passphrases didnt match", 1)
        return
    }

    err := WriteHashFromPassphrase(firstPassphrase)
    if err != nil {
		common.Kill("Error while writing hash file", 1)
        return
    }

    updateTime(TIMEOUT_SEC)
}

func askPassphrase() string {
	fmt.Print("Input passphrase: ")
	inputPassphrase := common.SecureGetPassword()
    return inputPassphrase
}

func isPassphrasesMatch(firstPassphrase string, secondPassphrase string) bool {
    return firstPassphrase == secondPassphrase
}

func getRestOfTime() int {
	restOfTime, _ := os.ReadFile(TIME_FILE)
	i, _ := strconv.Atoi(string(restOfTime))
	return i
}

func updateTime(seconds int) error {
	return os.WriteFile(TIME_FILE, []byte(calculateMomentOfExpire(int64(seconds))), 0600)
}

func calculateMomentOfExpire(timeout int64) string {
	now := time.Now()
    return fmt.Sprintf("%d", now.Unix() + timeout)
}

func initDeadmanSwitch() {
    cmd := exec.Command("touch", "/home/thek4n/DEADMAN")
    if err := cmd.Run(); err != nil {
        fmt.Println("Error: ", err)
    }

	log.Printf("Deadman Switch EXECUTED!")
	os.Exit(0)
}

func timeout() {
	for {
        checkPeriod := 15
        sleepSeconds(checkPeriod)

		if isExpire() {
			initDeadmanSwitch()
			break
		}
	}
}

func sleepSeconds(seconds int) {
    time.Sleep(time.Duration(seconds) * time.Second)
}

func isExpire() bool {
    return getRestOfTime() < int(time.Now().Unix())
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

		isValidHash, err := CheckHash(string(buf[:readLen]))
		if isValidHash {
			updateTime(TIMEOUT_SEC)
            log.Print("Extended until: " + getCurTime())
			conn.Write([]byte("Extended until: " + getCurTime()))
		} else {
			conn.Write([]byte("Declined, expires at: " + getCurTime()))
		}
		conn.Close()
		break
	}
}

func main() {
	os.Remove(common.SOCKET_FILE)
	syscall.Unlink(common.SOCKET_FILE)

	command := parseCommand()

	switch command {
	case "run":
		go timeout()

		listener, _ := net.Listen("unix", common.SOCKET_FILE)
		os.Open(common.SOCKET_FILE)
		os.Chown(common.SOCKET_FILE, 0, 1015)
		os.Chmod(common.SOCKET_FILE, 0660)
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
	syscall.Unlink(common.SOCKET_FILE)
	os.Remove(common.SOCKET_FILE)
}