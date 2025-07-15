package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"DSN"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	Path   string `mapstructure:"path"`
}

func LoadConfig() (*Config, error) {
	// -- default config --
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.DSN", "data/grampus.db")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.path", "./logs")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.grampus")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Not found config file! All use default config")
		} else {
			return nil, err
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
