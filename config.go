package gonnect

import (
	"io"

	"github.com/spf13/viper"
)

type Config struct {
	Development EnvironmentConfiguration
	Production  EnvironmentConfiguration
}

type EnvironmentConfiguration struct {
	Port    int
	BaseUrl string
	Store   StoreConfiguration
}

type StoreConfiguration struct {
	Type        string
	DatabaseUrl string
}

func NewConfig(configFile io.Reader) (error, *Config) {
	runtimeViper := viper.New()
	runtimeViper.SetConfigType("json") // or viper.SetConfigType("YAML")
	config := &Config{}

	err := runtimeViper.ReadConfig(configFile)
	if err != nil {
		return err, nil
	}
	err = runtimeViper.Unmarshal(config)
	if err != nil {
		return err, nil
	}

	return nil, config
}
