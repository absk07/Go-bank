package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBDriver string `mapstructure:"DB_DRIVER"`
	DBUri    string `mapstructure:"DB_URI"`
	Port     string `mapstructure:"PORT"`
	Secret string `mapstructure:"SECRET"`
}

func LoadConfig() (config Config, err error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
