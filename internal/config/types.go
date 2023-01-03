package config

import "WireguardManager/internal/core"

// Config root
type Configuration struct {
	App     AppConf
	Develop DevelopConf
}

// App config
type AppConf struct {
	LaunchMode LaunchMode `koanf:"launchmode" validate:"required,oneof=default develop"`

	Host struct {
		Ip           string `koanf:"ip" validate:"required,ipv4"`
		NetInterface string `koanf:"netinterface" validate:"required,min=1,max=15"`
	} `koanf:"host" validate:"required"`

	Wireguard struct {
		Port      int `koanf:"port" validate:"required,min=1,max=65535"`
		PeerLimit int `koanf:"peerlimit" validate:"required,min=1,max=65535"`
	} `koanf:"wireguard" validate:"required"`

	RestApi struct {
		Port    int    `koanf:"port" validate:"required,min=1,max=65535"`
		GinMode string `koanf:"ginmode" validate:"required,oneof=release debug test"`
	} `koanf:"restapi" validate:"required"`

	Database struct {
		Path string `koanf:"path" validate:"required"`
	} `koanf:"database" validate:"required"`

	Logging struct {
		FolderPath   string `koanf:"folderpath" validate:"required"`
		FileLevel    string `koanf:"filelevel" validate:"required"`
		ConsoleLevel string `koanf:"consolelevel" validate:"required"`
	} `koanf:"logging" validate:"required"`
}

type LaunchMode string

const (
	Develop LaunchMode = "develop"
	Default LaunchMode = "default"
)

// Dev configurationyy
type DevelopConf struct {
	Services struct {
		Server struct {
			Privatekey string `koanf:"privatekey" validate:"required,base64"`
			Port       int    `koanf:"privatekey" validate:"required,min=1,max=65535"`
		} `koanf:"server" validate:"required"`

		Peer struct {
			PeersToCreate []core.CreatePeer `koanf:"peers" validete:"omitempty,dive"`
		}
	}
}
