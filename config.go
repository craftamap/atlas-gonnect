package gonnect

import (
	"errors"
	"io"

	"github.com/spf13/viper"
)

type Config struct {
	CurrentProfile string
	Profiles       map[string]Profile
}

type Profile struct {
	Port            int
	BaseUrl         string
	Store           StoreConfiguration
	EnvironmentName string
}

type StoreConfiguration struct {
	Type        string
	DatabaseUrl string
}

func NewConfig(configFile io.Reader) (*Profile, string, error) {
	// TODO: I dont like this approach rn, we should at least store the
	// CurrentEnvironment so the programmers can interact with
	LOG.Info("Initializing Configuration")

	runtimeViper := viper.New()
	runtimeViper.SetDefault("CurrentProfile", "dev")
	runtimeViper.BindEnv("CurrentProfile", "GONNECT_PROFILE")
	runtimeViper.SetConfigType("json")
	config := &Config{}

	err := runtimeViper.ReadConfig(configFile)
	if err != nil {
		return nil, "", err
	}
	err = runtimeViper.Unmarshal(config)
	if err != nil {
		return nil, "", err
	}

	if config.CurrentProfile == "" {
		return nil, "", errors.New("No Profile selected; Set CurrentProfile in the config file or set GONNECT_PROFILE")
	}

	if profile, ok := config.Profiles[config.CurrentProfile]; !ok {
		return nil, "", errors.New("Profile not found!")
	} else {
		return &profile, config.CurrentProfile, nil
	}
}
