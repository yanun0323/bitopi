package main

import (
	"bitopi/intrernal/app"
	"bitopi/pkg/config"

	"github.com/yanun0323/pkg/logs"
)

var (
	l *logs.Logger
)

func main() {
	l = logs.New("bito_pi", 2)
	if err := config.Init("config"); err != nil {
		l.Fatalf("init config failed %s", err)
		return
	}

	app.Run(l)
}
