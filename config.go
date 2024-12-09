package utils

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type BaseConfig struct {
	Listen string
	Mode   string
}

func (b BaseConfig) Release() bool {
	return b.Mode == "release"
}

type Config[T any] interface {
	Release() bool
}

func LoadConfig[T Config[T]]() (*T, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	var obj T
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&obj); err != nil {
		return nil, err
	}

	{
		ho := &slog.HandlerOptions{
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.MessageKey {
					a.Key = "message"
				}
				return a
			},
		}
		if obj.Release() {
			slog.SetDefault(
				slog.New(slog.NewJSONHandler(os.Stderr, ho)),
			)
		} else {
			slog.SetDefault(
				slog.New(slog.NewTextHandler(os.Stderr, ho)),
			)
		}
	}

	return &obj, nil
}
