package main

import (
	"bitopi/internal/app"

	"github.com/yanun0323/pkg/config"
)

func main() {

	if err := config.Init("config", true, "./config", "../config", "../../config"); err != nil {
		panic("init config failed")
	}

	app.Run()
}
