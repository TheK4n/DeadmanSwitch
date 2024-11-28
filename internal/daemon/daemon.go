package daemon

import (
	"os"
	"strconv"
	"time"
	"log"

	"github.com/thek4n/DeadmanSwitch/internal/switcher"
)


type Daemon struct {
	Timeout int
	TimeFile string
	Switcher switcher.Switcher
}

func (d Daemon) Run() {
	checkPeriod := 15
	for {
		sleepSeconds(checkPeriod)

		momentOfExpiration, err := d.GetMomentOfExpiration()
		if err != nil {
			log.Printf("error: %s", err.Error())
			continue
		}

		if d.isExpire(momentOfExpiration) {
			d.Switcher.Execute()
			break
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