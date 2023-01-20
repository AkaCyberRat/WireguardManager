package config

type Configuration struct {
	Host struct {
		Ip string `koanf:"ip" validate:"required,ipv4"`
	} `koanf:"host" validate:"required"`

	Wireguard struct {
		Port      int `koanf:"port" validate:"required,min=1,max=65535"`
		PeerLimit int `koanf:"peerlimit" validate:"required,min=1,max=65535"`
	} `koanf:"wireguard" validate:"required"`

	RestApi struct {
		Port    int    `koanf:"port" validate:"required,min=1,max=65535"`
		GinMode string `koanf:"ginmode" validate:"required,oneof=release debug test"`

		Ssl struct {
			CrtPath *string `koanf:"crtpath" validate:"required_with=KeyPath"`
			KeyPath *string `koanf:"keypath" validate:"required_with=CrtPath"`
		} `koanf:"ssl" validate:"omitempty"`
	} `koanf:"restapi" validate:"required"`

	Database struct {
		FilePath string `koanf:"filepath" validate:"required"`
	} `koanf:"database" validate:"required"`

	Logging struct {
		FilePath     string `koanf:"filepath" validate:"required"`
		FileLevel    string `koanf:"filelevel" validate:"required"`
		ConsoleLevel string `koanf:"consolelevel" validate:"required"`
	} `koanf:"logging" validate:"required"`

	Jwt struct {
		HS256SecretKey     *string `koanf:"hs256secretkey" validate:"omitempty"`
		RS256PublicKeyPath *string `koanf:"rs256publickeypath" validate:"omitempty"`
	} `koanf:"jwt" validate:"omitempty"`
}
