package zinc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Index    string `json:"index" yaml:"index"`

	client *http.Client
}

func (c *Config) copy() *Config {
	return &Config{
		Host:     c.Host,
		User:     c.User,
		Password: c.Password,
		Index:    c.Index,
		client:   c.client,
	}
}

type With func(c *Config)

func (c *Config) With(ws ...With) *Config {
	n := c.copy()
	for _, v := range ws {
		v(n)
	}
	return n
}

func WithIndex(index string) func(*Config) {
	return func(c *Config) {
		c.Index = index
	}
}

func WithClient(client *http.Client) func(*Config) {
	return func(c *Config) {
		c.client = client
	}
}

func (c *Config) PushCheck(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return c.PushStrCheck(string(data))
}

func (c *Config) Push(obj any) {
	if err := c.PushCheck(obj); err != nil {
		os.Stderr.Write([]byte(err.Error()))
	}
}

func (c *Config) PushStr(data string) {
	if err := c.PushStrCheck(data); err != nil {
		os.Stderr.Write([]byte(err.Error()))
	}
}

func (c *Config) PushStrCheck(data string) error {
	req, err := http.NewRequest("POST", c.Host+"/api/"+c.Index+"/_doc", strings.NewReader(data))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.User, c.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("push failed code[%d]:%s", resp.StatusCode, string(data))
		fmt.Println("[ZINC]", err)
		return err
	}
	return nil
}

func (c *Config) Write(p []byte) (n int, err error) {
	return len(p), c.PushStrCheck(string(p))
}
