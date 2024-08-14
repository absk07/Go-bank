package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver              string        `mapstructure:"DB_DRIVER"`
	DBSource              string        `mapstructure:"DB_SOURCE"`
	DBMigrationURL        string        `mapstructure:"DB_MIGRATION_URL"`
	HTTP_Port             string        `mapstructure:"HTTP_PORT"`
	GRPC_Port             string        `mapstructure:"GRPC_PORT"`
	Secret                string        `mapstructure:"SECRET"`
	TokenDuration         time.Duration `mapstructure:"TOKEN_DURATION"`
	RefereshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
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
