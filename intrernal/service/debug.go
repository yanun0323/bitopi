package service

import (
	"bitopi/intrernal/util"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Service) DebugService(c echo.Context) error {
	pure, err := time.Parse("20060102", "20221125")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, util.NewErrorResponse("bad", err))
	}
	local, err := time.ParseInLocation("20060102", "20221125", time.Local)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, util.NewErrorResponse("bad", err))
	}
	localFixed, err := time.ParseInLocation("20060102", "20221125", time.FixedZone("UTC+8", 8*60*60))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, util.NewErrorResponse("bad", err))
	}
	localUTC, err := time.ParseInLocation("20060102", "20221125", time.UTC)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, util.NewErrorResponse("bad", err))
	}

	return c.JSON(http.StatusOK, util.NewDataResponse("good", struct {
		Pure       time.Time `json:"pure"`
		Local      time.Time `json:"local"`
		LocalFixed time.Time `json:"local_fixed"`
		LocalUTC   time.Time `json:"local_utc"`
		Now        time.Time `json:"now"`
		NowLocal   time.Time `json:"now_local"`
	}{
		Pure:       pure,
		Local:      local,
		LocalFixed: localFixed,
		LocalUTC:   localUTC,
		Now:        time.Now(),
		NowLocal:   time.Now().Local(),
	}))
}
