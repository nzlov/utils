package zinc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type Config struct {
	Host         string `json:"host" yaml:"host"`
	User         string `json:"user" yaml:"user"`
	Password     string `json:"password" yaml:"password"`
	Index        string `json:"index" yaml:"index"`
	Cache        int    `json:"cache" yaml:"cache"`
	CacheTimeout int    `json:"cacheTimeout" yaml:"cacheTimeout"`

	client *http.Client
	ch     chan string
}

func (c *Config) copy() *Config {
	return &Config{
		Host:     c.Host,
		User:     c.User,
		Password: c.Password,
		Index:    c.Index,
		client:   c.client,
		ch:       c.ch,
	}
}

type With func(c *Config)

func (c *Config) With(ws ...With) *Config {
	n := c.copy()
	for _, v := range ws {
		v(n)
	}
	if n.ch == nil {
		WithCache(c.Cache)(n)
	}
	if n.client == nil {
		WithClient(http.DefaultClient)(n)
	}
	if n.Index == "" {
		WithIndex("test")(n)
	}
	if n.CacheTimeout <= 0 {
		WithCacheTimeout(100)(n)
	}
	return n
}

func WithIndex(index string) func(*Config) {
	return func(c *Config) {
		c.Index = index
	}
}

func WithCache(cache int) func(*Config) {
	return func(c *Config) {
		if cache < 1 {
			cache = 10
		}
		c.ch = make(chan string, cache)
		c.Cache = cache
		go c.push()
	}
}

func WithClient(client *http.Client) func(*Config) {
	return func(c *Config) {
		c.client = client
	}
}

func WithCacheTimeout(t int) func(*Config) {
	return func(c *Config) {
		c.CacheTimeout = t
	}
}

func (c *Config) push() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
		go c.push()
	}()

	var buffer bytes.Buffer

	i := 0
	for {
		select {
		case data := <-c.ch:
			buffer.WriteString(data)
			i++
			if i > c.Cache/2 {
				c.zincPush(buffer.String())
				i = 0
				buffer.Reset()
			}
		case <-time.After(time.Duration(c.CacheTimeout) * time.Millisecond):
			if i > 0 {
				c.zincPush(buffer.String())
				i = 0
				buffer.Reset()
			}
		}
	}
}

func (c *Config) zincPush(data string) {
	req, err := http.NewRequest("POST", c.Host+"/api/"+c.Index+"/_multi", strings.NewReader(data))
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	req.SetBasicAuth(c.User, c.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("push failed code[%d]:%s", resp.StatusCode, string(data))
		os.Stderr.WriteString("[ZINC]" + err.Error())
	}
}

func (c *Config) PushCheck(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c.PushStrCheck(string(data))
	return nil
}

func (c *Config) Push(obj any) {
	c.PushCheck(obj)
}

func (c *Config) PushStr(data string) {
	c.PushStrCheck(data)
}

func (c *Config) PushStrCheck(data string) {
	c.ch <- data
}

func (c *Config) Write(p []byte) (n int, err error) {
	c.PushStrCheck(string(p))
	return len(p), nil
}
