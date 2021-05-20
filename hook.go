package logrus_influxdb

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
)

// InfluxDBHook delivers logs to an InfluxDB cluster.
// as logrus does not exploit io.Closer interface,
// we're unable to gracefully close InfluxDB and flush batch buffer
type InfluxDBHook struct {
	client influxdb.Client
	config Config
	//use only via getter
	organizationID *string
}

// newInfluxDBClient
// return influxdb.Client
func newInfluxDBClient(oConfig TConfigClient) (influxdb.Client, error) {

	if nil != oConfig.Options {
		return influxdb.NewClientWithOptions(oConfig.Url(), oConfig.Token, oConfig.Options), nil
	}

	return influxdb.NewClient(oConfig.Url(), oConfig.Token), nil
}

// obtainClient obtains existing client provided with optional arguments
// or creates a new one
func obtainClient(oConfig TConfigClient, slClient []influxdb.Client) (oClient influxdb.Client, err error) {
	switch len(slClient) {
	case 1:
		oClient = slClient[0]
	case 0:
		if oClient, err = newInfluxDBClient(oConfig); nil != err {
			err = fmt.Errorf("NewInfluxDB: Error creating InfluxDB Client, %v", err)
			return
		}
	default:
		err = fmt.Errorf("NewInfluxDB: Error creating InfluxDB Client, %d is too many influxdb slClient", len(slClient))
		return
	}

	return
}

// NewInfluxDB returns a new InfluxDBHook
// influxdb.Client is optional
func NewInfluxDB(oConfig Config, slClient ...influxdb.Client) (pHook *InfluxDBHook, err error) {

	if err = oConfig.Check(); nil != err {
		return
	}

	var oClient influxdb.Client
	if oClient, err = obtainClient(oConfig.TConfigClient, slClient); nil != err {
		return
	}

	if err = oConfig.CheckInflux(oClient); nil != err {
		return
	}

	pHook = &InfluxDBHook{
		client: oClient,
		config: oConfig,
	}

	if err = pHook.autoCreateBucket(); nil != err {
		pHook = nil
		return
	}

	return
}

// autoCreateBucket tries to detect if the bucket exists and if not,
// automatically creates one if allowed by TConfigWriteAPI.RequireBucket.
func (pHook *InfluxDBHook) autoCreateBucket() (err error) {

	//due to lack of error types in influx API, it is unable to distinguish missing bucket error from any other
	//so let's pretend we can create it anyway
	_, err = pHook.client.BucketsAPI().FindBucketByName(context.Background(), pHook.config.Bucket)

	if nil != err && !pHook.config.RequireBucket {
		pBucket := &domain.Bucket{
			Name:  pHook.config.Bucket,
			OrgID: pHook.orgID(),
		}
		_, err = pHook.client.BucketsAPI().CreateBucket(context.Background(), pBucket)
	}

	return
}

//return organization id if organization with name from config exists; otherwise returns empty string
func (pHook *InfluxDBHook) orgID() *string {
	if nil == pHook.organizationID {
		pOrg, _ := pHook.client.OrganizationsAPI().FindOrganizationByName(context.Background(), pHook.config.Organization)
		if nil != pOrg {
			pHook.organizationID = pOrg.Id
		}
	}

	return pHook.organizationID
}

func (pHook *InfluxDBHook) writeAPI() api.WriteAPI {
	return pHook.client.WriteAPI(pHook.config.OrgBucket())
}

func (pHook *InfluxDBHook) writeAPIBlocking() api.WriteAPIBlocking {
	return pHook.client.WriteAPIBlocking(pHook.config.OrgBucket())
}

func (pHook *InfluxDBHook) Close() error {
	pHook.writeAPI().Flush()
	pHook.client.Close()

	return nil
}

// Levels are the available logging levels.
func (pHook *InfluxDBHook) Levels() []logrus.Level {
	return pHook.config.Level
}

// Fire asynchronously adds a new InfluxDB point based off of Logrus entry to a batch
// if TConfigClient.Immediate is set, writes synchronously (development mode only)
func (pHook *InfluxDBHook) Fire(pEntry *logrus.Entry) error {
	pPoint, err := pHook.config.populatePoint(pEntry)
	if nil != err {
		return err
	}
	if pHook.config.Immediate {
		return pHook.writeAPIBlocking().WritePoint(context.Background(), pPoint)
	}
	pHook.writeAPI().WritePoint(pPoint)

	return nil
}
