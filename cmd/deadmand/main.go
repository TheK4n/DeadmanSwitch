package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"path/filepath"

	"github.com/thek4n/DeadmanSwitch/internal/passphrases"
	"github.com/thek4n/DeadmanSwitch/internal/daemon"
	"github.com/thek4n/DeadmanSwitch/internal/switcher"
)

var PREFIX = os.Getenv("HOME") + "/.local/deadman"
var TIME_FILE = PREFIX + "/time"
var HASH_FILE = PREFIX + "/hash"
const SOCKET_FILE = "/tmp/deadman.sock"

const DEADMAN_TIMEOUT_VARIABLE_NAME = "DEADMAN_TIMEOUT"
const DEADMAN_COMMAND_VARIABLE_NAME = "DEADMAN_COMMAND"


func main() {
	handleSignals()

	_, programName := filepath.Split(os.Args[0])
	if len(os.Args) < 2 {
		die(1, "Usage: %s <COMMAND>", programName)
	}

	switch os.Args[1] {
	case "run":
		runDaemon()
	case "init":
		initialSetup()
	case "--help":
		fmt.Println("run\ninit")
		os.Exit(0)
	default:
		die(1, "'%s' is not a %s command.", os.Args[1], programName)
	}
}

func handleSignals() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		s := <-sigChan
		cleanupSocket(SOCKET_FILE)

		fmt.Println(s)
		os.Exit(137)
	}()
}

func runDaemon() {
	deadmanCommand := os.Getenv(DEADMAN_COMMAND_VARIABLE_NAME)
	if deadmanCommand == "" {
		die(1, "%s variable not set", DEADMAN_COMMAND_VARIABLE_NAME)
	}

	timeoutSec, getTimeoutErr := getTimeoutSec()
	if getTimeoutErr != nil {
		die(1, "%s is invalid", DEADMAN_TIMEOUT_VARIABLE_NAME)
	}

	d := daemon.Daemon{
		Timeout: timeoutSec,
		TimeFile: TIME_FILE,
		Switcher: switcher.ShellCommandSwitcher{Command: deadmanCommand},
	}

	go d.Run()

	listener, listenErr := net.Listen("unix", SOCKET_FILE)
	if listenErr != nil {
		cleanupSocket(SOCKET_FILE)
		die(1, "listen socket %s", listenErr.Error())
		return
	}

	chmodErr := os.Chmod(SOCKET_FILE, 0660)
	if chmodErr != nil {
		cleanupSocket(SOCKET_FILE)
		die(1, "chmod socket %s", chmodErr.Error())
		return
	}

	log.Printf("Server starts")
	momentOfExpiration, _ := d.GetMomentOfExpiration()
	log.Printf("Expires at: %d", momentOfExpiration)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleClient(conn, d)
	}
}

func handleClient(conn net.Conn, d daemon.Daemon) {
	defer conn.Close()

	buf := make([]byte, 64)
	for {
		_, writeErr := conn.Write([]byte("Write passphrase:"))
		if writeErr != nil {
			log.Print("Error communicate via socket: " + writeErr.Error())
			conn.Close()
			cleanupSocket(SOCKET_FILE)
			break
		}

		readLen, readSocketErr := conn.Read(buf)
		if readSocketErr != nil {
			fmt.Println(readSocketErr.Error())
			conn.Close()
			cleanupSocket(SOCKET_FILE)
			break
		}

		messageFromClient := strings.Fields(string(buf[:readLen]))

		if len(messageFromClient) < 2 {
			momentOfExpiration, _ := d.GetMomentOfExpiration()
			conn.Write([]byte("Declined, expires at: " + string(momentOfExpiration)))
			conn.Close()
			cleanupSocket(SOCKET_FILE)
			break
		}

		command := messageFromClient[0]
		passphrase := messageFromClient[1]

		a, _ := os.ReadFile(HASH_FILE)
		isValidHash := passphrases.CheckHash(passphrase, string(a))

		if !isValidHash {
			momentOfExpiration, _ := d.GetMomentOfExpiration()
			conn.Write([]byte("Declined, expires at: " + string(momentOfExpiration)))
			conn.Close()

			cleanupSocket(SOCKET_FILE)
			break
		}

		switch command {
		case "extend":
			timeoutSec, _ := getTimeoutSec()

			updateExpireMomentErr := updateExpireMoment(timeoutSec)
			if updateExpireMomentErr != nil {
				fmt.Println("Update expire moment error" + updateExpireMomentErr.Error())
				conn.Close()

				cleanupSocket(SOCKET_FILE)
				break
			}

			momentOfExpiration, _ := d.GetMomentOfExpiration()
			log.Printf("Extended until: %d", momentOfExpiration)
			conn.Write([]byte("Extended until: " + string(momentOfExpiration)))

			conn.Close()

			cleanupSocket(SOCKET_FILE)
			break

		case "execute":
			res, err := d.Switcher.Execute()

			if err != nil {
				fmt.Println(err.Error())
			}

			fmt.Println(string(res))

			conn.Write([]byte("Executed"))
			conn.Close()

			cleanupSocket(SOCKET_FILE)
			break
		}

		momentOfExpiration, _ := d.GetMomentOfExpiration()
		conn.Write([]byte("Declined, expires at: " + string(momentOfExpiration)))

		conn.Close()

		cleanupSocket(SOCKET_FILE)
		break
	}
}
func cleanupSocket(socketfile string) {
   unlink_err := syscall.Unlink(socketfile)
   if unlink_err != nil {
       log.Printf("Unlink socket error: " + unlink_err.Error())
   }
}


func initialSetup() {
	timeoutSec, getTimeoutErr := getTimeoutSec()
	if getTimeoutErr != nil {
		die(1, "%s is invalid", DEADMAN_COMMAND_VARIABLE_NAME)
	}

	firstPassphrase, secondPassphrase := askPassphraseTwice()

	if !isPassphrasesMatch(firstPassphrase, secondPassphrase) {
		die(1, "passphrases didnt match")
	}

	err := WriteHashFromPassphrase(firstPassphrase, HASH_FILE)
	if err != nil {
		die(1, "writing hash file")
	}

	updateExpireMomentErr := updateExpireMoment(timeoutSec)
	if updateExpireMomentErr != nil {
		die(1, "writing time file")
		return
	}
}

func askPassphraseTwice() (string, string) {
	fmt.Print("Input passphrase: ")
	firstPassphrase := secureGetPassword()

	fmt.Print("Repeat passphrase: ")
	secondPassphrase := secureGetPassword()

	return firstPassphrase, secondPassphrase
}

func isPassphrasesMatch(firstPassphrase string, secondPassphrase string) bool {
	return firstPassphrase == secondPassphrase
}

func getTimeoutSec() (int64, error) {
	return strconv.ParseInt(os.Getenv(DEADMAN_TIMEOUT_VARIABLE_NAME), 10, 64)
}


func WriteHashFromPassphrase(passphrase string, hashFile string) error {
	return writeHash(passphrases.HashSaltPassphrase(passphrase), hashFile)
}

func writeHash(hash string, hashFile string) error {
	hashfile_dir := filepath.Dir(hashFile)
	err := os.MkdirAll(hashfile_dir, 0700)

	if err != nil {
		return err
	}
	return os.WriteFile(hashFile, []byte(hash), 0600)
}

func updateExpireMoment(seconds int64) error {
	expireMoment := calculateExpireMoment(seconds)
	return os.WriteFile(TIME_FILE, []byte(strconv.FormatInt(expireMoment, 10)), 0600)
}

func calculateExpireMoment(timeout int64) int64 {
	now := time.Now()
	return now.Unix() + timeout
}

func secureGetPassword() string {
	var input string
	fmt.Print("\033[8m")
	fmt.Scanf("%s", &input)
	fmt.Print("\033[28m")
	return input
}

func die(code int, format string, a ...any) {
	_, programName := filepath.Split(os.Args[0])

	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s: error: ", programName) + format, a...)
	os.Exit(code)
}