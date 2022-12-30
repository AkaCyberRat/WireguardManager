package config

type Config struct {
	ServerIP          string `mapstructure:"SERVER_IP" validate:"required,ipv4"`
	WgPort            int    `mapstructure:"WG_PORT" validate:"required,min=1,max=65535"`
	WgPeerLimit       int    `mapstructure:"PEER_LIMIT" validate:"required,min=1,max=65535"`
	ApiPort           int    `mapstructure:"API_PORT" validate:"required,min=1,max=65535"`
	Interface         string `mapstructure:"INTERFACE" validate:"required,min=1,max=15"`
	UseSSL            bool   `mapstructure:"USE_SSL"`
	UseTrafficControl bool   `mapstructure:"USE_TRAFFIC_CONTROL"`
	UseTestConfig     bool   `mapstructure:"USE_TESTING_CONFIG"`
	LogLevelFile      string `mapstructure:"FILE_LOGLEVEL" validate:"required"`
	LogLevelConsole   string `mapstructure:"CONSOLE_LOGLEVEL" validate:"required"`
}

type TestingConfig struct {
	WgPrivateKey           string `mapstructure:"WG_PRIVATE_KEY" validate:"required,base64"`
	FirstPeerPublicKey     string `mapstructure:"FIRST_PEER_PUBLIC_KEY" validate:"required,base64"`
	FirstPeerPresharedKey  string `mapstructure:"FIRST_PEER_PRESHARED_KEY" validate:"required,base64"`
	FirstPeerDownloadSpeed int    `mapstructure:"FIRST_PEER_DOWNLOAD_SPEED" validate:"required,min=1,max=200"`
	FirstPeerUploadSpeed   int    `mapstructure:"FIRST_PEER_UPLOAD_SPEED" validate:"required,min=1,max=200"`
}
