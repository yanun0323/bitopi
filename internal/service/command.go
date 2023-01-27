package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type SlackCommand struct {
	Service
	SlackCommandOption
}

type SlackCommandOption struct {
	Name  string
	Token string
}

func NewCommand(svc Service, opt SlackCommandOption) SlackCommand {
	return SlackCommand{
		Service:            svc,
		SlackCommandOption: opt,
	}
}

func (svc *SlackCommand) Handler(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		svc.l.Errorf("parse form, %+v\n", errors.WithStack(err))
		return svc.ok(c, "internal error")
	}

	payload := c.Request().PostForm
	callbackUrl := util.NewUrl(payload.Get("response_url"))
	userID := payload.Get("user_id")
	svc.l.Debug("user: ", userID, ", from channel: ", payload.Get("channel_id"))
	directChannel := svc.getDirectChannel(userID)
	if len(directChannel) == 0 {
		return svc.sendCommandReply(callbackUrl, "找不到私人訊息頻道")
	}

	fromChannel := payload.Get("channel_id")
	if directChannel != fromChannel {
		return svc.sendCommandReply(callbackUrl, "指令只能在應用程式私訊使用，請到應用程式對話輸入指令")
	}

	// FIXME: Add administrator validator
	// valid, err := svc.repo.IsAdmin(userID)
	// if err != nil {
	// 	return svc.sendCommandReply(callbackUrl, fmt.Sprintf("無法驗證管理員, %s", err))
	// }

	commands := svc.splitRequestText(payload.Get("text"))
	if len(commands) == 0 {
		return svc.sendCommandReply(callbackUrl, "需要輸入指令，執行 `help` 取得更多資訊")
	}

	// TODO: Handle command
	switch cmd := strings.ToLower(commands[0]); cmd {
	case "clear":
		if err := svc.cmdClearAllMsgReply(directChannel); err != nil {
			return svc.sendCommandReply(callbackUrl, fmt.Sprintf("執行指令行為 `%s` 錯誤，%+v", cmd, err))
		}
		return nil

	default:
		return svc.sendCommandReply(callbackUrl, fmt.Sprintf("找不到指令行為 `%s`，執行 `help` 取得更多資訊", cmd))
	}
}

func (svc *SlackCommand) sendCommandReply(url util.Url, msg string) error {
	notifier := util.NewSlackNotifier(svc.Token)
	res, code, err := notifier.Send(svc.ctx, http.MethodPost, url, &util.GeneralMsg{Text: msg})
	if err != nil {
		fmt.Printf("bot send, %s\n", err)
	}
	svc.l.Debug("code: ", code)
	svc.l.Debug("res: ", string(res))
	return nil
}

func (svc *SlackCommand) splitRequestText(text string) []string {
	var sp, br byte = ' ', '\n'
	result := []string{}

	queue := []byte{}
	for i := range text {
		if text[i] != sp && text[i] != br {
			queue = append(queue, text[i])
			continue
		}
		if len(queue) > 0 {
			result = append(result, string(queue))
			queue = []byte{}
		}
	}
	if len(queue) > 0 {
		result = append(result, string(queue))
	}
	return result
}

func (svc *SlackCommand) getDirectChannel(userID string) string {
	buf, _, err := util.HttpRequest(util.HttpRequestOption{
		Method:       http.MethodPost,
		Url:          "https://slack.com/api/conversations.open?users=" + userID,
		Token:        svc.Token,
		IsUrlencoded: true,
	})
	if err != nil {
		svc.l.Errorf("send http request error, %+v", err)
		return ""
	}

	c := &model.DirectChannel{}
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

func (svc *SlackCommand) cmdClearAllMsgReply(directChannel string) error {
	notifier := util.NewSlackNotifier(svc.Token)
	path := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s", directChannel)
	res, _, err := notifier.Send(svc.ctx, http.MethodGet, util.Url(path), &util.GeneralMsg{})
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	json.Unmarshal(res, &data)
	if err != nil {
		return err
	}

	if data["messages"] == nil {
		svc.l.Warn("empty messages")
		return nil
	}

	for _, msg := range data["messages"].([]interface{}) {
		ts := msg.(map[string]interface{})["ts"].(string)
		res, _, err := notifier.Send(svc.ctx, http.MethodPost, util.Url("https://slack.com/api/chat.delete"), &util.SlackSimpleMsg{
			Channel:   directChannel,
			Timestamp: ts,
		})

		if err != nil {
			return err
		}

		svc.l.Debugf("%s", string(res))
	}

	return nil
}
