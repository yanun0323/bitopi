package app

import (
	"bitopi/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

const (
	_tokenHeaderKey = "TOKEN"
)

func tokenValidator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		valid := viper.GetString("admin.token")
		token := c.Request().Header.Get(_tokenHeaderKey)
		if len(valid) == 0 {
			return service.ErrorResponse(c, http.StatusInternalServerError, "service token not set")
		}

		if !strings.EqualFold(valid, token) {
			return service.ErrorResponse(c, http.StatusBadRequest, "invalid token")
		}

		return next(c)
	}
}
