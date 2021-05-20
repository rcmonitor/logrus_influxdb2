package main

import (
	"fmt"
	"github.com/rcmonitor/logrus_influxdb2"
	"github.com/sirupsen/logrus"
	"time"
)

// Major client and API config is taken from .env provided
// use example.env as a template
//
// missing bucket will trigger error
func main() {

	log := logrus.New()

	oConfig := logrus_influxdb.Config{}
	//from now on,
	//it's everyone for himself
	oConfig.RequireBucket = true
	oConfig.Bucket = "non-existent_bucket"
	//let's populate defaults from .env
	if err := oConfig.Check(); nil != err {
		fmt.Printf("config check failed: %s \n", err)
		return
	}

	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	log.Debugf("let's explore at %s", time.Now().String())
}
