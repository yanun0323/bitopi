package app

import (
	"bitopi/intrernal/service"
	"log"
	"net/http"

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
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Msg string `json:"message"`
		}{
			Msg: "OK",
		})
	})
	e.POST("/backend-maid", svc.MaidBotHandler, m...)
	e.POST("/backend-maid/command", svc.MaidCommandHandler, m...)

	e.Start(":8001")
}
