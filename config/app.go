package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	DBDriver          string `mapstructure:"DB_DRIVER"`
	DBSource1         string `mapstructure:"DB_SOURCE_1"`
	DBSource2         string `mapstructure:"DB_SOURCE_2"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.SetDefault("HTTP_SERVER_ADDRESS", "0.0.0.0:10082")
	viper.SetDefault("DB_DRIVER", "mysql")
	viper.SetDefault("DB_SOURCE_1", "root@tcp(127.0.0.1:3306)/stingray_local?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("DB_SOURCE_2", "")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
