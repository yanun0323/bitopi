package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (svc *Service) ok(c echo.Context, i ...interface{}) error {
	svc.l.Info("OK")
	if len(i) > 0 {
		return c.JSON(http.StatusOK, i[0])
	}
	return c.JSON(http.StatusOK, nil)
}

func (svc *Service) getDirectChannel(userID, token string) string {
	buf, _, err := util.HttpRequest(util.HttpRequestOption{
		Method:       http.MethodPost,
		Url:          "https://slack.com/api/conversations.open?users=" + userID,
		Token:        token,
		IsUrlencoded: true,
	})
	if err != nil {
		svc.l.Errorf("send http request error, %+v", err)
		return ""
	}

	c := &model.SlackDirectChannel{}
	if err := json.Unmarshal(buf, c); err != nil {
		svc.l.Errorf("unmarshal json response error, %+v", err)
		return ""
	}

	if !c.OK {
		svc.l.Errorf("invalid request, response:\n%s", string(buf))
		return ""
	}

	svc.l.Debugf("success get direct channel id %s", c.Channel.ID)
	return c.Channel.ID
}
