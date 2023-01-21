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
	"gopkg.in/yaml.v3"
)

const (
	ConfigFilepath = "./configs/config.json"
	EnvPrefix      = "APP_"
)

func LoadConfiguration() (*Configuration, error) {
	//
	// Create configurator instanse
	//
	configurator := koanf.New(".")

	//
	// (First layer) Set default values
	//
	err := configurator.Load(confmap.Provider(map[string]interface{}{
		"host.ip":              "127.0.0.1",
		"wireguard.port":       51820,
		"wireguard.peerlimit":  100,
		"restapi.port":         5000,
		"restapi.ginmode":      "release",
		"dataBase.filepath":    "./db/service.db",
		"logging.filepath":     "./log/logs.txt",
		"logging.consolelevel": "info",
		"logging.filelevel":    "debug",
	}, "."), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to load default configuration: %v", err.Error())
	}

	//
	// (Second layer) Load config.json
	//
	err = configurator.Load(file.Provider(ConfigFilepath), json.Parser())
	if err == nil {
		logrus.Info("User configuration has been read.")

	} else if os.IsNotExist(err) {
		logrus.Warn("User configuration file wasn't found.")

	} else {
		return nil, fmt.Errorf("failed to read user configuration file: %v", err.Error())
	}

	//
	// (Third layer) Load environment variables
	//
	err = configurator.Load(env.Provider(EnvPrefix, ".", convertEnvVarName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err.Error())
	}
	logrus.Info("Env variables have been read.")

	//
	// Unmarshal app config
	//
	var config Configuration
	err = configurator.Unmarshal("", &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err.Error())
	}

	//
	// Validate app config
	//
	err = validator.New().Struct(config)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err.Error())
	}

	showConfiguration(config)

	return &config, nil
}

func convertEnvVarName(s string) string {

	return strings.Replace(
		strings.ToLower(strings.TrimPrefix(s, EnvPrefix)), "_", ".", -1)
}

func showConfiguration(conf Configuration) {
	bytes, _ := yaml.Marshal(conf)
	strs := strings.Split(string(bytes), "\n")

	logrus.Infof("Current configuration:\n")
	for _, v := range strs[:len(strs)-2] {
		logrus.Infof("%s", v)
	}
	logrus.Infof("%s\n", strs[len(strs)-2])
}
