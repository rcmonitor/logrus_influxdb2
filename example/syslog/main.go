package main

import (
	"fmt"
	"github.com/rcmonitor/logrus_influxdb2"
	"github.com/sirupsen/logrus"
	"time"
)

// major client and API config is taken from .env provided
// use example.env as a template
//
// using syslog-style
func main() {
	log := logrus.New()
	pConfigSysLog := &logrus_influxdb.TConfigSyslog{
		Facility:     "local0",
		FacilityCode: 16,
		AppName:      "TheBestAppEver",
		Version:      "v0.0.42",
	}
	oConfig := logrus_influxdb.Config{
		TConfigSyslog: pConfigSysLog,
	}
	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	log.Warnf("Gigafail at %s", time.Now().Format("15:04:05"))
	time.Sleep(2 * time.Second)
}
