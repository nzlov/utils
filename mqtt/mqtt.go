package mqtt

import mqtt "github.com/eclipse/paho.mqtt.golang"

type MqttConfig struct {
	mqtt mqtt.Client

	Host     string `json:"host"   yaml:"host"   mapstructure:"host"`
	User     string `json:"user"   yaml:"user"   mapstructure:"user"`
	Password string `json:"password"   yaml:"password"   mapstructure:"password"`
	ClientID string `json:"client"   yaml:"client"   mapstructure:"client"`
}

func (c *MqttConfig) Mqtt() mqtt.Client {
	if c.mqtt != nil {
		return c.mqtt
	}
	m, err := c.Open()
	if err != nil {
		panic(err)
	}
	return m
}

func (c *MqttConfig) Open() (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.Host)
	opts.SetClientID(c.ClientID)
	if c.User != "" {
		opts.SetUsername(c.User)
		opts.SetPassword(c.Password)
	}

	c.mqtt = mqtt.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c.mqtt, nil
}
