package main

import (
	"bitopi/intrernal/app"
	"bitopi/pkg/config"
)

func main() {
	if err := config.Init("config"); err != nil {
		panic("init config failed")
	}

	app.Run()
}
