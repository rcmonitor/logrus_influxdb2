# InfluxDB2 Hook for Logrus

Bug-reports, issues and pull request are kindly appreciated

Kinda, based on [abramovic/logrus_influxdb](https://github.com/abramovic/logrus_influxdb)

- [Examples](https://github.com/rmonitor/logrus_influxdb2/tree/master/examples)
- [Logrus](https://github.com/Sirupsen/logrus)
- [InfluxDB](https://influxdb.com)

#### [Contributors](https://github.com/rcmonitor/logrus_influxdb2/graphs/contributors)

## Basic usage

```go
package main

import (
    "fmt"
    
    "github.com/rcmonitor/logrus_influxdb2"
    "github.com/sirupsen/logrus"
)

func main() {
    log := logrus.New()
    
    oConfig := logrus_influxdb.Config{}
    oConfig.Organization = "org_test"
    oConfig.Token = "GmT--n80qIOL_Ph2bJ15J9ZMvVZhk0AjnxXu1CWnzyKs_TGHmgKRUgiiEUyGEFx4ppZY31A=="
    oConfig.Port = 48086

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

```

## Behind the scenes

#### Bucket and Measurement Handling

When passing an empty string for the InfluxDB bucket name, it is set to "logrus".
When passing an empty string for the InfluxDB measurement name, it is set to "logrus".

When initializing the hook we attempt to first see if the bucket exists. If not, by default we try to create it for your automagically.
This behaviour can be changed by setting `Config.TConfigWriteAPI.RequireBucket` to `true`

Measurement can be dynamically overwritten using `logrus.WithField/s()`

#### Message Field

We will insert your message into InfluxDB with the field `message` so please make sure not to use that name with your Logrus fields or else it will be overwritten.

#### Special Fields

Some logrus fields have a special meaning in this hook, these are `logger`  (taken from [Sentry Hook](https://github.com/evalphobia/logrus_sentry)).

- `logger` is the part of the application which is logging the event
