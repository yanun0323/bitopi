package service

import (
	"github.com/labstack/echo/v4"
)

type SlackAction struct {
	Name string
	SlackBot
}

type SlackActionOption struct {
}

func NewAction(name string, bot SlackBot) SlackAction {
	return SlackAction{
		Name:     name,
		SlackBot: bot,
	}
}

func (svc *SlackAction) Handler(c echo.Context) error {
	return svc.ok(c, svc.eventCallbackResponse(c))
}

func (svc *SlackAction) eventCallbackResponse(c echo.Context) interface{} {
	return struct {
		DeleteOriginal bool `json:"delete_original"`
	}{
		DeleteOriginal: true,
	}
}
