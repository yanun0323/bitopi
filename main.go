package main

import (
	"bitopi/intrernal/app"
	"bitopi/pkg/config"

	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/logs"
)

var (
	l *logs.Logger
)

func main() {
	if err := config.Init("config"); err != nil {
		panic("init config failed")
	}
	l = logs.New("bito_pi", uint8(viper.GetInt("log.level")))

	app.Run(l)
}
