package gonnect

import (
	"errors"
	"io"

	"github.com/spf13/viper"
)

type Config struct {
	CurrentEnvironment string
	Development        EnvironmentConfiguration
	Production         EnvironmentConfiguration
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

func NewConfig(configFile io.Reader) (*EnvironmentConfiguration, error) {
	runtimeViper := viper.New()
	runtimeViper.SetEnvPrefix("gonnect")
	runtimeViper.BindEnv("CurrentEnvironment")
	runtimeViper.SetConfigType("json") // or viper.SetConfigType("YAML")
	config := &Config{}

	err := runtimeViper.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}
	err = runtimeViper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	if config.CurrentEnvironment == "development" {
		return &config.Development, nil
	} else if config.CurrentEnvironment == "production" {
		return &config.Production, nil
	} else {
		return nil, errors.New("No Environment set")
	}

}
