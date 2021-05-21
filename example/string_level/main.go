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
// use influx only for required logging levels
// explicitly provide list of level titles to log
func main() {
	log := logrus.New()
	oConfig := logrus_influxdb.Config{}
	oConfig.LevelTitle = []string{"info", "error"}
	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	strFormat := "15:04:05"
	//this one is not going to be written in influx
	log.Warnf("take care at %s", time.Now().Format(strFormat))
	//these two should land in influx
	log.Infof("going on at %s", time.Now().Format(strFormat))
	log.Errorf("error at %s", time.Now().Format(strFormat))
	time.Sleep(2 * time.Second)
}
