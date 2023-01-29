package service

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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
	action, err := svc.parseAction(c)
	if err != nil {
		svc.l.Errorf("parse action error, %+v", err)
		return nil
	}

	switch action {
	case "resend":
		svc.l.Info("execute resend")
	case "delete":
		svc.l.Info("execute delete")
	}

	return struct {
		DeleteOriginal bool `json:"delete_original"`
	}{
		DeleteOriginal: true,
	}
}

func (svc *SlackAction) parseAction(c echo.Context) (string, error) {
	data := map[string]string{}
	if err := c.Bind(&data); err != nil {
		return "", err
	}

	if len(data["payload"]) == 0 {
		return "", errors.New("empty payload content")
	}

	params := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data["payload"]), &params); err != nil {
		return "", err
	}

	if params["actions"] == nil {
		return "", errors.New("empty actions content")
	}

	actions, ok := params["actions"].([]interface{})
	if !ok {
		return "", errors.New("transfer actions type error")
	}

	value, exist := actions[0].(map[string]interface{})["value"]
	if !exist {
		return "", errors.New("empty actions value")
	}

	svc.l.Debug("action value: ", value)
	return value.(string), nil
}

func (svc *SlackAction) parseRequest(c echo.Context) error {
	data := map[string]string{}
	if err := c.Bind(&data); err != nil {
		return err
	}

	if len(data["payload"]) == 0 {
		return errors.New("empty payload")
	}

	params := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data["payload"]), &params); err != nil {
		return err
	}

	for k, v := range params {
		if k == "original_message" {
			svc.l.Debug("--- original_message ---")
			for kk, vv := range v.(map[string]interface{}) {
				svc.l.Debug(kk, ": ", vv)
			}
			svc.l.Debug("------")
			continue
		}

		svc.l.Debug(k, ": ", v)
	}

	return nil
}
