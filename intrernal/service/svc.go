package service

import (
	"bitopi/intrernal/domain"
	"bitopi/intrernal/repository"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Service struct {
	repo domain.IRepository
	l    *log.Logger
}

func New() (Service, error) {
	repo, err := repository.NewRepo()
	if err != nil {
		return Service{}, err
	}
	return Service{
		repo: repo,
	}, nil
}

func ok(c echo.Context, i ...interface{}) error {
	fmt.Println("OK")
	if len(i) > 0 {
		return c.JSON(http.StatusOK, i[0])
	}
	return c.JSON(http.StatusOK, nil)
}
