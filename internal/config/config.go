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

func NewConfig() (*Configuration, error) {
	//
	// Create configurator instanse
	//
	configurator := koanf.New(".")

	//
	// (First layer) Set default values
	//
	err := configurator.Load(confmap.Provider(map[string]interface{}{
		"app.launchmode": "default",

		"app.host.ip":           "127.0.0.1",
		"app.host.netinterface": "eth0",

		"app.wireguard.port":      51820,
		"app.wireguard.peerlimit": 100,

		"app.restapi.port":    5000,
		"app.restapi.ginmode": "release",

		"app.dataBase.path": "/app/db/service.db",

		"app.logging.folderpath":   "/app/logs/",
		"app.logging.consolelevel": "info",
		"app.logging.filelevel":    "error",
	}, "."), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to load default configuration: %v", err.Error())
	}

	//
	// (Second layer) Load config.json
	//
	err = configurator.Load(file.Provider("/app/configs/config.json"), json.Parser())
	if err == nil {
		logrus.Info("User configuration has been read.")

	} else if os.IsNotExist(err) {
		logrus.Warn("User configuration file wasn't found. (Default settings applied)")

	} else {
		return nil, fmt.Errorf("failed to read user configuration file: %v", err.Error())
	}

	//
	// (Third layer) Load environment variables
	//
	err = configurator.Load(env.Provider("APP_", ".", convertEnvVarName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err.Error())
	}

	//
	// Unmarshal app config
	//
	var config Configuration
	err = configurator.Unmarshal("app", &config.App)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err.Error())
	}

	//
	// Validate app config
	//
	err = validator.New().Struct(config.App)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err.Error())
	}

	//
	// Load dev conig if 'develop' launch mode
	//
	if config.App.LaunchMode == Develop {
		config.Develop, err = loadDevConfig()
		return &config, err
	}

	return &config, nil
}

func loadDevConfig() (DevelopConf, error) {
	var devConf DevelopConf

	k := koanf.New(".")

	err := k.Load(file.Provider("/app/configs/config.json.dev"), json.Parser())
	if err != nil {
		return DevelopConf{}, fmt.Errorf("failed to read dev config file: %v", err.Error())
	}

	logrus.Warn("Development configuraton has been read.")

	err = k.Unmarshal("dev", &devConf)
	if err != nil {
		return DevelopConf{}, fmt.Errorf("failed to unmarshal dev config: %v", err.Error())
	}

	err = validator.New().Struct(devConf)
	if err != nil {
		return DevelopConf{}, fmt.Errorf("dev config validation failed: %v", err.Error())
	}

	return devConf, nil
}

func convertEnvVarName(s string) string {

	return strings.Replace(
		strings.TrimPrefix(s, "APP_"), "_", ".", -1)
}
