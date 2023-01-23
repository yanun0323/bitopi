package main

import (
	"bitopi/internal/app"
	"bitopi/pkg/config"
)

func main() {
	if err := config.Init("config"); err != nil {
		panic("init config failed")
	}

	app.Run()
}
