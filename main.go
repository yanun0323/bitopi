package main

import (
	"bitopi/intrernal/app"
	"bitopi/pkg/config"
	"log"
)

var (
	l *log.Logger
)

func main() {
	l = log.Default()
	if err := config.Init("config"); err != nil {
		l.Fatalf("init config failed %s", err)
		return
	}

	app.Run(l)
}
