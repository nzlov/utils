package redis

import "github.com/redis/go-redis/v9"

type Config struct {
	redis *redis.Client

	Addr     string `json:"addr"     yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db"       yaml:"db"`
}

func (c *Config) Redis() *redis.Client {
	if c.redis != nil {
		return c.redis
	}
	c.redis = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})
	return c.redis
}
