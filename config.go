package logrus_influxdb

import (
	"context"
	"fmt"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

var (
	defaultHost string = "localhost"
	defaultPort int    = 8086

	defaultBucket      string = "logrus"
	defaultMeasurement string = "logrus"

	defaultLevel = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
)

type Config struct {
	TConfigLog
	TConfigClient
	TConfigWriteAPI
	*TConfigSyslog
	Tag []string
}

type TConfigClient struct {
	Host     string
	Port     int
	Token    string
	UseHTTPS bool
	Options  *influxdb.Options
	//use only for debugging purposes
	//forces write to influx without buffer
	Immediate bool
}

func (pConfig *TConfigClient) defaultsClient() (err error) {
	if "" == pConfig.Token {
		pConfig.Token = os.Getenv("LOGRUS_INFLUX_TOKEN")
		if "" == pConfig.Token {
			return fmt.Errorf("token is obligatory for influxdb client initiation")
		}
	}
	if "" == pConfig.Host {
		pConfig.Host = os.Getenv("LOGRUS_INFLUX_HOST")
		if "" == pConfig.Host {
			pConfig.Host = defaultHost
		}
	}
	if 0 == pConfig.Port {
		strPort := os.Getenv("LOGRUS_INFLUX_PORT")
		if pConfig.Port, err = strconv.Atoi(strPort); nil != err {
			return
		}
		if 0 == pConfig.Port {
			pConfig.Port = defaultPort
		}
	}

	return
}

// Url forms influx instance url based on
// schema, host and port
func (pConfig *TConfigClient) Url() string {
	strSchema := "http"
	if pConfig.UseHTTPS {
		strSchema += "s"
	}

	return fmt.Sprintf("%s://%s:%d", strSchema, pConfig.Host, pConfig.Port)
}

type TConfigWriteAPI struct {
	Organization string
	Bucket       string
	Measurement  string
	// if set to true only use bucket if it already exists
	// if set to false, bucket will be created automatically
	RequireBucket bool
}

func (pConfig *TConfigWriteAPI) OrgBucket() (string, string) {
	return pConfig.Organization, pConfig.Bucket
}

func (pConfig *TConfigWriteAPI) defaultsWriteAPI() {
	if "" == pConfig.Organization {
		pConfig.Organization = os.Getenv("LOGRUS_INFLUX_ORG")
	}
	if "" == pConfig.Bucket {
		if pConfig.Bucket = os.Getenv("LOGRUS_INFLUX_BUCKET"); "" == pConfig.Bucket {
			pConfig.Bucket = defaultBucket
		}
	}
	if "" == pConfig.Measurement {
		if pConfig.Measurement = os.Getenv("LOGRUS_INFLUX_MEASUREMENT"); "" == pConfig.Measurement {
			pConfig.Measurement = defaultMeasurement
		}
	}
}

type TConfigSyslog struct {
	Facility     string
	FacilityCode int
	AppName      string
	Version      string
}

type TConfigLog struct {
	//maximum level to log
	//if provided, every level up to MaxLevel including
	//considered worth logging
	MaxLevel string
	//list of levels to log
	//precedes MaxLevel and LevelTitle
	//if no levels provided using either MaxLevel, or Level list,
	//all levels considered worth logging
	Level []logrus.Level
	//list of level titles to log
	//precedes MaxLevel
	LevelTitle []string
}

// defaultsLog sets up all log levels to involve influx
// if none provided in config
func (pConfigLog *TConfigLog) defaultsLog() error {
	//levels defined with slice of logrus.Level
	if 0 < len(pConfigLog.Level) {
		return nil
	}
	//levels defined as slice of string literals
	if 0 < len(pConfigLog.LevelTitle) {
		for _, strLevel := range pConfigLog.LevelTitle {
			if intLevel, err := logrus.ParseLevel(strLevel); nil != err {
				return err
			}else{
				pConfigLog.Level = append(pConfigLog.Level, intLevel)
			}
		}
	}
	//levels defined as string literal of maximum loggable level
	if "" != pConfigLog.MaxLevel {
		intLevelOffset, err := logrus.ParseLevel(pConfigLog.MaxLevel)
		if nil != err { return err }
		//as soon as levels corresponds to offsets in defaultLevel array,
		//we can rely on offset itself
		pConfigLog.Level = defaultLevel[:intLevelOffset+ 1]
	}
	//no levels defined explicitly, let's fall back to defaults
	if 0 == len(pConfigLog.Level) {
		pConfigLog.Level = defaultLevel
	}

	return nil
}

// Check validates config
// sets defaults from either default variables
// or environment variables
// acts as a decorator for sub-configs population and validation
func (pConfig *Config) Check() error {
	if err := pConfig.defaultsLog(); nil != err { return err }
	pConfig.defaultsWriteAPI()
	return pConfig.defaultsClient()
}

// CheckInflux verifies that influx is set up correctly to process writes
// decorates specific functions that performs actual check-up
// @todo possibly, API write check should be added to verify token validity on operations in bucket
func (oConfig Config) CheckInflux(oClient influxdb.Client) error {
	if err := oConfig.checkHealth(oClient); nil != err {
		return err
	}
	return oConfig.checkOrg(oClient)
}

// checkOrg checks Organization existence
func (oConfig Config) checkOrg(oClient influxdb.Client) (err error) {
	var pOrganization *domain.Organization
	pOrganization, err = oClient.OrganizationsAPI().FindOrganizationByName(context.Background(), oConfig.Organization)
	if nil == err && nil == pOrganization {
		err = fmt.Errorf("%s: orgainzation not found", oConfig.Organization)
		return
	}

	return
}

// checkHealth validates influx endpoint availability
func (oConfig Config) checkHealth(oClient influxdb.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pCheck, err := oClient.Health(ctx); nil != err {
		return fmt.Errorf("influxdb health Check failed: %w", err)
	} else {
		if pCheck.Status != domain.HealthCheckStatusPass {
			return fmt.Errorf("influxdb not ready")
		}
	}

	return nil
}

func (oConfig Config) populatePoint(pEntry *logrus.Entry) (pPoint *write.Point, err error) {

	strMeasurement := oConfig.Measurement
	if result, ok := getTag(pEntry.Data, "measurement"); ok {
		strMeasurement = result
	}

	if "" == strMeasurement {
		err = fmt.Errorf("measurement name required")
		return
	}

	mTag := make(map[string]string)
	mField := make(map[string]interface{})

	if nil != oConfig.TConfigSyslog {

		var strHostName string
		strHostName, err = os.Hostname()

		if nil != err {
			return
		}

		strSeverity, intSeverityCode := parseSeverity(pEntry.Level.String())

		mTag["appname"] = oConfig.AppName
		mTag["facility"] = oConfig.Facility
		mTag["host"] = strHostName
		mTag["hostname"] = strHostName
		mTag["strSeverity"] = strSeverity

		mField["facility_code"] = oConfig.FacilityCode
		mField["message"] = pEntry.Message
		mField["procid"] = os.Getpid()
		mField["severity_code"] = intSeverityCode
		mField["timestamp"] = pEntry.Time.UnixNano()
		mField["version"] = oConfig.Version
	} else {
		// If passing a "message" field then it will be overridden by the entry Message
		pEntry.Data["message"] = pEntry.Message

		// Set the level of the entry
		mTag["level"] = pEntry.Level.String()
		// getAndDel and getAndDelRequest are taken from https://github.com/evalphobia/logrus_sentry
		if logger, ok := getTag(pEntry.Data, "logger"); ok {
			mTag["logger"] = logger
		}

		for k, v := range pEntry.Data {
			mField[k] = v
		}

		//migrate required mField to tag
		for _, strTagName := range oConfig.Tag {
			if strTagValue, ok := getTag(pEntry.Data, strTagName); ok {
				mTag[strTagName] = strTagValue
				delete(mField, strTagName)
			}
		}
	}

	pPoint = influxdb.NewPoint(strMeasurement, mTag, mField, pEntry.Time)

	return
}
