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
		if obj.Release() {
			slog.SetDefault(
				slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
					AddSource: true,
				})),
			)
		} else {
			slog.SetDefault(
				slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					AddSource: true,
				})),
			)
		}
	}

	return &obj, nil
}
