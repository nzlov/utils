package influx

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Config struct {
	ic influxdb2.Client

	Host   string `json:"host"   yaml:"host"   mapstructure:"host"`
	Token  string `json:"token"  yaml:"token"  mapstructure:"token"`
	Org    string `json:"org"    yaml:"org"    mapstructure:"org"`
	Bucket string `json:"bucket" yaml:"bucket" mapstructure:"bucket"`
}

func (c *Config) Influx() influxdb2.Client {
	if c.ic != nil {
		return c.ic
	}
	c.ic = influxdb2.NewClient(c.Host, c.Token)
	return c.ic
}

func (c *Config) Writer() api.WriteAPI {
	return c.Influx().WriteAPI(c.Org, c.Bucket)
}

func (c *Config) BucketWriter(b string) api.WriteAPI {
	return c.Influx().WriteAPI(c.Org, b)
}

func (c *Config) Query() api.QueryAPI {
	return c.Influx().QueryAPI(c.Org)
}

type ctxKey struct{}

var _ctxKey = ctxKey{}

func For(ctx context.Context) *Config {
	return ctx.Value(_ctxKey).(*Config)
}

func (c *Config) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, c)
}
