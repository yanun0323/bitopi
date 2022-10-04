package app

import (
	"bitopi/intrernal/service"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run(l *log.Logger) {

	e := echo.New()
	e.Logger.SetLevel(4)

	svc, err := service.New()
	if err != nil {
		l.Fatalf("create service failed %s", err)
		return
	}

	rateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
	m := []echo.MiddlewareFunc{rateLimiter}

	e.POST("/backend-maid", svc.MaidBotHandler, m...)
	e.POST("/backend-maid/command", svc.MaidCommandHandler, m...)

	e.Start(":8080")
}