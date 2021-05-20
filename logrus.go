package logrus_influxdb

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Try to return a field from logrus
// Taken from Sentry adapter (from https://github.com/evalphobia/logrus_sentry)
func getTag(d logrus.Fields, key string) (tag string, ok bool) {
	v, ok := d[key]
	if !ok {
		return "", false
	}
	switch vs := v.(type) {
	case fmt.Stringer:
		return vs.String(), true
	case string:
		return vs, true
	case byte:
		return string(vs), true
	case int:
		return strconv.FormatInt(int64(vs), 10), true
	case int32:
		return strconv.FormatInt(int64(vs), 10), true
	case int64:
		return strconv.FormatInt(vs, 10), true
	case uint:
		return strconv.FormatUint(uint64(vs), 10), true
	case uint32:
		return strconv.FormatUint(uint64(vs), 10), true
	case uint64:
		return strconv.FormatUint(vs, 10), true
	default:
		return "", false
	}
}

func parseSeverity(level string) (string, int) {
	switch level {
	case "panic":
		return "panic", 0
	case "fatal":
		return "crit", 2
	case "error":
		return "err", 3
	case "warning":
		return "warning", 4
	case "info":
		return "info", 6
	case "debug":
		return "debug", 7
	default:
		return "none", -1
	}
}
