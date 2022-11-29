package config

type _GeneralConfig struct {
	BaseConfig
	Language string
}

func (c *_GeneralConfig) Name() string {
	return "General"
}

var General = &_GeneralConfig{
	Language: "en",
}
