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
// measurement can be dynamically overwritten using logrus.WithField/s
func main() {
	log := logrus.New()
	oConfig := logrus_influxdb.Config{}
	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	strFormat := "15:04:05"

	log.WithField("measurement", "dynamic_measurement_1").
		Infof("passing line of code at %s", time.Now().Format(strFormat))

	oFields := logrus.Fields{
		"measurement":  "dynamic_measurement_2",
		"custom_field": 42,
	}
	log.WithFields(oFields).Warnf("beware of %s", time.Now().Format(strFormat))

	time.Sleep(2 * time.Second)
}
