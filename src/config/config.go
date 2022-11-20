package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var config Config

type Config struct {
	ServerIP  string `mapstructure:"SERVER_IP" validate:"required,ipv4"`
	WgPort    int    `mapstructure:"WG_PORT" validate:"required,min=0,max=65535"`
	PeerLimit int    `mapstructure:"PEER_LIMIT" validate:"required,min=0,max=65534"`

	ApiPort   int    `mapstructure:"API_PORT" validate:"required,min=1,max=65535"`
	Interface string `mapstructure:"INTERFACE" validate:"required,min=1,max=15"`

	UseTC  bool `mapstructure:"USE_TC"`
	UseSSL bool `mapstructure:"USE_SSL"`

	FileLogLevel    string `mapstructure:"FILE_LOGLEVEL" validate:"required"`
	ConsoleLogLevel string `mapstructure:"CONSOLE_LOGLEVEL" validate:"required"`
}

func Get() Config {
	return config
}

func Load() {
	viperSetDefaults()
	viper.AddConfigPath("/app/config")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		logrus.Fatal("Failed to read config: %v", err.Error())
	}

	config = Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		logrus.Fatal("Failed to unmarshal config: %v", err.Error())
	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		logrus.Fatal("Failed to validate config: %v", err.Error())
	}
}

func viperSetDefaults() {
	viper.SetDefault("SERVER_IP", "127.0.0.1")
	viper.SetDefault("WG_PORT", 51820)
	viper.SetDefault("API_PORT", 5000)
	viper.SetDefault("PEER_LIMIT", 100)
	viper.SetDefault("INTERFACE", "eth0")
	viper.SetDefault("USE_TC", true)
	viper.SetDefault("USE_SSL", false)
	viper.SetDefault("FILE_LOGLEVEL", "debug")
	viper.SetDefault("CONSOLE_LOGLEVEL", "debug")
}
