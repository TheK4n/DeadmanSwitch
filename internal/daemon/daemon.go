package daemon

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/thek4n/DeadmanSwitch/internal/switcher"
)


type Daemon struct {
	Timeout int64
	TimeFile string
	Switcher switcher.Switcher
}

func (d Daemon) Run() {
	checkPeriod := 3
	for {
		sleepSeconds(checkPeriod)
		momentOfExpiration, err := d.GetMomentOfExpiration()
		if err != nil {
			log.Printf("error: %s", err.Error())
			continue
		}

		if d.isExpire(momentOfExpiration) {
			res, err := d.Switcher.Execute()

			if err != nil {
				fmt.Println(err.Error())
			}

			fmt.Println(string(res))
			fmt.Println("DEADMAN executed!")
			os.Exit(1)
		}
	}
}

func (d Daemon) isExpire(momentOfExpiration int64) bool {
	now := time.Now().Unix()
	return momentOfExpiration < now
}

func (d Daemon) GetMomentOfExpiration() (int64, error) {
	momentOfExpirationString, err := os.ReadFile(d.TimeFile)
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseInt(string(momentOfExpirationString), 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func sleepSeconds(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}