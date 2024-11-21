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
    "os/signal"

	common "../common"
)

const TIMEOUT_SEC int = 60

var PREFIX = os.Getenv("HOME") + "/.local/deadman"
var TIME_FILE = PREFIX + "/time"
var HASH_FILE = PREFIX + "/hash"


func main() {
    sigChan := make(chan os.Signal, 1)

    signal.Notify(
        sigChan,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
    )

    go func() {
        s := <-sigChan
        cleanupSocket(common.SOCKET_FILE)

        fmt.Println(s)
        os.Exit(137)
    }()

	command := parseCommand(os.Args)
    handleCommand(command)
}

func parseCommand(args[] string) string {
	if len(args) < 2 {
		common.Kill("Usage: "+args[0]+" COMMAND", 1)
	}
	command := args[1]
	return command
}

func handleCommand(command string) {
	switch command {
        case "run":
            runDaemon()
        case "init":
            initialSetup()
        default:
            common.Kill("'"+os.Args[1]+"' is not a "+os.Args[0]+" command.", 1)
	}
}

func runDaemon() {
    go timeoutDaemon()

    listener, listen_err := net.Listen("unix", common.SOCKET_FILE)
    if listen_err != nil {
        cleanupSocket(common.SOCKET_FILE)
        common.Kill("Error listen socket file" + listen_err.Error(), 1)
        return
    }

    chmod_err := os.Chmod(common.SOCKET_FILE, 0660)
    if chmod_err != nil {
        cleanupSocket(common.SOCKET_FILE)
        common.Kill("Error chmod socket file" + chmod_err.Error(), 1)
        return
    }

    log.Printf("Server starts")
    log.Printf("Expires at: " + getExpireMoment())

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        go handleClient(conn)
    }
}

func cleanupSocket(socketfile string) {
    unlink_err := syscall.Unlink(socketfile)
    if unlink_err != nil {
        log.Printf("Unlink socket error: " + unlink_err.Error())
    }
}

func timeoutDaemon() {
	for {
        checkPeriod := 15
        sleepSeconds(checkPeriod)

		if isExpire() {
			initDeadmanSwitch()
			break
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 32)
	for {
		conn.Write([]byte("Write passphrase:"))

		readLen, readSocketErr := conn.Read(buf)
		if readSocketErr != nil {
			fmt.Println(readSocketErr)
			break
		}

		isValidHash, _ := CheckHash(string(buf[:readLen]))

		if isValidHash {
			updateExpireMoment(TIMEOUT_SEC)
            log.Print("Extended until: " + getExpireMoment())
			conn.Write([]byte("Extended until: " + getExpireMoment()))
		} else {
			conn.Write([]byte("Declined, expires at: " + getExpireMoment()))
		}

		conn.Close()
		break
	}
}

func writeHash(hash string) error {
    hashfile_dir := filepath.Dir(HASH_FILE)
    err := os.MkdirAll(hashfile_dir, 0700)

    if err != nil {
        return err
    }
	return os.WriteFile(HASH_FILE, []byte(hash), 0600)
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

    updateExpireMomentErr := updateExpireMoment(TIMEOUT_SEC)
    if updateExpireMomentErr != nil {
		common.Kill("Error while writing time file", 1)
        return
    }
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

func updateExpireMoment(seconds int) error {
	return os.WriteFile(TIME_FILE, []byte(calculateExpireMoment(int64(seconds))), 0600)
}

func calculateExpireMoment(timeout int64) string {
	now := time.Now()
    return fmt.Sprintf("%d", now.Unix() + timeout)
}

func initDeadmanSwitch() {
    cmd := exec.Command("touch", "/home/thek4n/DEADMAN")
    if err := cmd.Run(); err != nil {
        fmt.Println("Error: ", err)
        cleanupSocket(common.SOCKET_FILE)
        os.Exit(0)
    }

	log.Printf("Deadman Switch EXECUTED!")

    cleanupSocket(common.SOCKET_FILE)
	os.Exit(0)
}

func sleepSeconds(seconds int) {
    time.Sleep(time.Duration(seconds) * time.Second)
}

func isExpire() bool {
    return getRestOfTime() < int(time.Now().Unix())
}

func getExpireMoment() string {
	return fmt.Sprintf("%s", time.Unix(int64(getRestOfTime()), 0))
}