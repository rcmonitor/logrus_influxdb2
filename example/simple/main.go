package main

import (
	"fmt"
	"github.com/rcmonitor/logrus_influxdb2"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

// the simplest usage possible
// all the required configuration settings
// are set up implicitly
//
// bucket and measurement is set to 'logrus' if not provided by environment
// and created automatically
// host is set to localhost
func main() {
	log := logrus.New()
	oConfig := logrus_influxdb.Config{}
	//oConfig.Organization = "org_test"
	//oConfig.Token = "GmT--n80qIOL_Ph2bJ15J9ZMvVZhk0AjnxXu1CWnzyKs_TGHmgKRUgiiEUyGEFx4ppZY31A=="
	//oConfig.Port = 48086

	oConfig.Organization = os.Getenv("LOGRUS_INFLUX_ORG")
	oConfig.Port, _ = strconv.Atoi(os.Getenv("LOGRUS_INFLUX_PORT"))
	oConfig.Token = os.Getenv("LOGRUS_INFLUX_TOKEN")

	hook, err := logrus_influxdb.NewInfluxDB(oConfig)
	if nil != err {
		fmt.Printf("error creating influxdb hook: %v", err)
		return
	}
	log.Hooks.Add(hook)
	log.Errorf("superfail")
	//timeout should be greater than default flush interval for influxdb (1 sec)
	time.Sleep(2 * time.Second)
}
