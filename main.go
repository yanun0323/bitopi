package main

import (
	"bitopi/internal/app"
	"bitopi/pkg/infra"
)

func main() {
	if err := infra.Init("config"); err != nil {
		panic("init config failed")
	}

	app.Run()
}
