package service

import (
	"bitopi/intrernal/domain"
	"bitopi/intrernal/repository"
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yanun0323/pkg/logs"
)

type Service struct {
	repo domain.IRepository
	l    *logs.Logger
}

func New() (Service, error) {
	repo, err := repository.NewRepo()
	if err != nil {
		return Service{}, err
	}
	return Service{
		repo: repo,
		l:    logs.Get(context.Background()),
	}, nil
}

func ok(c echo.Context, i ...interface{}) error {
	logs.Get(context.Background()).Info("OK")
	if len(i) > 0 {
		return c.JSON(http.StatusOK, i[0])
	}
	return c.JSON(http.StatusOK, nil)
}
