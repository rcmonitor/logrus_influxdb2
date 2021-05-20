package main

import (
	"fmt"
	"github.com/rcmonitor/logrus_influxdb2"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// major client and API config is taken from .env provided
// use example.env as a template
//
// explicit exit function can perform graceful shutdown of influx
// on sudden exit
func main() {
	log := logrus.New()

	log.ExitFunc = func(intCode int) {
		//as we're unable to get level of last fired log entry,
		//we're forced to trigger io.Close on every level available
		for _, slHook := range log.Hooks {
			for _, iHook := range slHook {
				if iCloser, ok := iHook.(io.Closer); ok {
					_ = iCloser.Close()
				}
			}
		}
		os.Exit(intCode)
	}

	oConfig := logrus_influxdb.Config{}
	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	//fatal runs exit func and performs os.Exit()
	log.Fatalf("time to panic: %s", time.Now().Format("15:04:05"))

	//// the same effect can be achieved using logrus.Exit()
	//log.Errorf("bad error")
	//log.Exit(0)
}
