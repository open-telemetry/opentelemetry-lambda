package solarwindsapmsettingsextension

import "time"

type Config struct {
	Endpoint string        `mapstructure:"endpoint"`
	Key      string        `mapstructure:"key"`
	Interval time.Duration `mapstructure:"interval"`
}
