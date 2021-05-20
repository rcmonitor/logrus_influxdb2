package main

import (
	"fmt"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rcmonitor/logrus_influxdb2"
	"github.com/sirupsen/logrus"
	"time"
)

// major client and API config is taken from .env provided
// use example.env as a template
//
// providing a self-created client gives a possibility to gracefully shut down influx
// and flush buffer without worrying about logging near exit point
func main() {

	log := logrus.New()
	pOptions := influxdb.DefaultOptions().AddDefaultTag("hook_name", "test_hook")

	oConfig := logrus_influxdb.Config{}
	//let's populate defaults from .env
	if err := oConfig.Check(); nil != err {
		fmt.Printf("config check failed: %s \n", err)
		return
	}

	oClient := influxdb.NewClientWithOptions(oConfig.Url(), oConfig.Token, pOptions)

	hook, err := logrus_influxdb.NewInfluxDB(oConfig, oClient)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	log.Infof("doing great at %s", time.Now().String())

	oClient.Close()
}
