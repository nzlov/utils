package influx

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxConfig struct {
	ic influxdb2.Client

	Host   string `json:"host"   yaml:"host"   mapstructure:"host"`
	Token  string `json:"token"  yaml:"token"  mapstructure:"token"`
	Org    string `json:"org"    yaml:"org"    mapstructure:"org"`
	Bucket string `json:"bucket" yaml:"bucket" mapstructure:"bucket"`
}

func (c *InfluxConfig) Influx() influxdb2.Client {
	if c.ic != nil {
		return c.ic
	}
	c.ic = influxdb2.NewClient(c.Host, c.Token)
	return c.ic
}

func (c *InfluxConfig) InfluxWriter() api.WriteAPI {
	return c.Influx().WriteAPI(c.Org, c.Bucket)
}

func (c *InfluxConfig) InfluxBucketWriter(b string) api.WriteAPI {
	return c.Influx().WriteAPI(c.Org, b)
}

func (c *InfluxConfig) InfluxQuery() api.QueryAPI {
	return c.Influx().QueryAPI(c.Org)
}
