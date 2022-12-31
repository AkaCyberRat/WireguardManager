package config

type Config struct {
	HostIp           string `koanf:"Host.Ip" validate:"required,ipv4"`
	HostNetInterface string `koanf:"Host.NetInterface" validate:"required,min=1,max=15"`

	WireguardPort      int `koanf:"Wireguard.Port" validate:"required,min=1,max=65535"`
	WireguardPeerLimit int `koanf:"Wireguard.PeerLimit" validate:"required,min=1,max=65535"`

	ApiPort  int  `koanf:"Api.Port" validate:"required,min=1,max=65535"`
	ApiUseTC bool `koanf:"Api.UseTC"`

	DataBasePath string `koanf:"DataBase.Path" validate:"required"`

	LoggingFilePath     string `koanf:"Logging.Path" validate:"required"`
	LoggingFileLevel    string `koanf:"Logging.ConsoleLevel" validate:"required"`
	LoggingConsoleLevel string `koanf:"Logging.FileLevel" validate:"required"`
}

type TestingConfig struct {
	WgPrivateKey           string `mapstructure:"WG_PRIVATE_KEY" validate:"required,base64"`
	FirstPeerPublicKey     string `mapstructure:"FIRST_PEER_PUBLIC_KEY" validate:"required,base64"`
	FirstPeerPresharedKey  string `mapstructure:"FIRST_PEER_PRESHARED_KEY" validate:"required,base64"`
	FirstPeerDownloadSpeed int    `mapstructure:"FIRST_PEER_DOWNLOAD_SPEED" validate:"required,min=1,max=200"`
	FirstPeerUploadSpeed   int    `mapstructure:"FIRST_PEER_UPLOAD_SPEED" validate:"required,min=1,max=200"`
}
