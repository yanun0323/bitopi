package app

import (
	"bitopi/intrernal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/logs"
)

func Run() {
	l := logs.New("bito_pi", uint8(viper.GetInt("log.level")))
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
	}, m...)
	e.GET("/debug", svc.DebugService, m...)

	e.POST("/pm", svc.PMBotHandler, m...)
	e.POST("/rails-hi", svc.RailsBotHandler, m...)
	e.POST("/devops-bro", svc.DevopsBotHandler, m...)
	e.POST("/backend-maid", svc.MaidBotHandler, m...)
	e.POST("/backend-maid/command", svc.MaidCommandHandler, m...)

	e.Start(":8001")
}
