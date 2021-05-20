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
// client is fine-tuned using influxdb.Options
func main() {

	log := logrus.New()

	//batch size of 1 makes it write *almost* immediately
	pInfluxOptions := influxdb.DefaultOptions().SetBatchSize(1).SetFlushInterval(500)

	oConfigClient := logrus_influxdb.TConfigClient{
		Options: pInfluxOptions,
	}
	oConfig := logrus_influxdb.Config{
		TConfigClient: oConfigClient,
	}
	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	log.Errorf("superfail at %s", time.Now().String())
	time.Sleep(1 * time.Millisecond)
}
