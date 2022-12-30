package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewConfig() (*Config, error) {
	configViper := viper.New()

	// Set defaults
	configViper.SetDefault("SERVER_IP", "127.0.0.1")
	configViper.SetDefault("WG_PORT", 51820)
	configViper.SetDefault("API_PORT", 5000)
	configViper.SetDefault("PEER_LIMIT", 100)
	configViper.SetDefault("INTERFACE", "eth0")
	configViper.SetDefault("USE_TC", true)
	configViper.SetDefault("USE_SSL", false)
	configViper.SetDefault("FILE_LOGLEVEL", "debug")
	configViper.SetDefault("CONSOLE_LOGLEVEL", "warning")

	// Read user config
	configViper.AddConfigPath("/app")
	configViper.SetConfigName("config")
	configViper.SetConfigType("json")
	configViper.AutomaticEnv()

	err := configViper.ReadInConfig()
	if err == nil {
		logrus.Info("Config was read")
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		logrus.Info("Config was not found and skipped. (Used default settings)")
	} else {
		return nil, fmt.Errorf("failed to read config: %v", err.Error())
	}

	// Read dev config
	configViper.AddConfigPath("/app")
	configViper.SetConfigName("config-dev")
	configViper.SetConfigType("json")

	err = configViper.MergeInConfig()
	if err == nil {
		logrus.Warn("Dev config was read")
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok { // Check if file exist and reading failed
		logrus.Debug("Dev config not used")
	} else {
		return nil, fmt.Errorf("failed to read dev config: %v", err.Error())
	}

	// Unmarshal config
	config := Config{}
	err = configViper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err.Error())
	}

	// Validate config
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %v", err.Error())
	}

	return &config, nil
}
