package config

import (
	"bitopi/internal/util"

	"github.com/spf13/viper"
)

func Init(configName string) error {
	name := configName
	if len(configName) > 0 {
		name = "config"
	}
	viper.AddConfigPath(util.GetAbsPath("config"))
	viper.AddConfigPath(util.GetAbsPath())
	viper.AutomaticEnv()
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")

	return viper.ReadInConfig()
}
