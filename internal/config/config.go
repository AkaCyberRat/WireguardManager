package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"
)

func NewConfig() (*Config, error) {
	//
	// Create koanf configurator instanse
	//
	k := koanf.New(".")

	//
	// (First config layer) Set default values
	//
	err := k.Load(confmap.Provider(map[string]interface{}{
		"Host.Ip":           "127.0.0.1",
		"Host.NetInterface": "eth0",

		"Wireguard.Port":      51820,
		"Wireguard.PeerLimit": 100,

		"Api.Port":  5000,
		"Api.UseTC": true,

		"DataBase.Path": "/app/db/service.db",

		"Logging.FilePath":     "/app/logs/",
		"Logging.ConsoleLevel": "info",
		"Logging.FileLevel":    "error",
	}, "."), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to load default configuration: %v", err.Error())
	}

	//
	// (Second config layer) Load config.json
	//
	err = k.Load(file.Provider("/app/configs/config.json"), json.Parser())
	if err == nil {
		logrus.Info("User configuration has been read.")

	} else if os.IsNotExist(err) {
		logrus.Warn("User configuration file wasn't found. (Default settings applied)")

	} else {
		return nil, fmt.Errorf("failed to read user configuration file: %v", err.Error())
	}

	//
	// (Third config layer) Load development config.json.dev
	//
	err = k.Load(file.Provider("/app/configs/config.json.dev"), json.Parser())
	if err == nil {
		logrus.Warn("Development configuraton has been read.")

	} else if os.IsNotExist(err) { // Check if file exist and reading failed
		logrus.Debug("User configuration file wasn't found.")

	} else {
		return nil, fmt.Errorf("failed to read development configuration file: %v", err.Error())
	}

	//
	// (Fourth config layer) Load environment variables
	//
	err = k.Load(env.Provider("APP_", ".", convertEnvVarName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err.Error())
	}

	//
	// Unmarshal data from configurator to struct
	//
	var config Config
	err = k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err.Error())
	}

	//
	// Validate config
	//
	err = validator.New().Struct(config)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err.Error())
	}

	return &config, nil
}

func convertEnvVarName(s string) string {

	return strings.Replace(
		strings.TrimPrefix(s, "APP_"), "_", ".", -1)
}
