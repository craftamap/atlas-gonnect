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
	Port            int
	BaseUrl         string
	Store           StoreConfiguration
	EnvironmentName string
}

type StoreConfiguration struct {
	Type        string
	DatabaseUrl string
}

func NewConfig(configFile io.Reader) (*EnvironmentConfiguration, error) {
	// TODO: I dont like this approach rn, we should at least store the
	// CurrentEnvironment so the programmers can interact with
	LOG.Info("Initializing Configuration")

	runtimeViper := viper.New()
	runtimeViper.SetDefault("CurrentEnvironment", "development")
	runtimeViper.SetEnvPrefix("gonnect")
	runtimeViper.BindEnv("CurrentEnvironment")
	runtimeViper.SetConfigType("json")
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
		LOG.Info("development configuration initialized")
		config.Development.EnvironmentName = "development"
		return &config.Development, nil
	} else if config.CurrentEnvironment == "production" {
		LOG.Info("configuration initialized")
		config.Production.EnvironmentName = "production"
		return &config.Production, nil
	} else {
		return nil, errors.New("No Environment set")
	}
}
