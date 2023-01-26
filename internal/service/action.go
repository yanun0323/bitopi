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

func (act *SlackAction) Handler(c echo.Context) error {
	return act.ok(c, act.eventCallbackResponse(c))
}

func (act *SlackAction) eventCallbackResponse(c echo.Context) interface{} {
	return struct {
		DeleteOriginal bool `json:"delete_original"`
	}{
		DeleteOriginal: true,
	}
}
