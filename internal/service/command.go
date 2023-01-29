package service

import (
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	_commandClearNotification = "\n\n> `⌘ + R` 可清除此訊息"
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
	callbackUrl := util.Url(payload.Get("response_url"))
	userID := payload.Get("user_id")
	svc.l.Debug("user: ", userID, ", from channel: ", payload.Get("channel_id"))
	directChannel := svc.getDirectChannel(userID, svc.Token)
	if len(directChannel) == 0 {
		return svc.sendCommandReply(callbackUrl, "找不到私人訊息頻道")
	}

	fromChannel := payload.Get("channel_id")
	if directChannel != fromChannel {
		return svc.sendCommandReply(callbackUrl, "指令只能在應用程式私訊使用，請到應用程式對話輸入指令")
	}

	commands := svc.splitRequestText(payload.Get("text"))
	if len(commands) == 0 {
		return svc.sendCommandReply(callbackUrl, "需要輸入指令，執行 `help` 取得更多資訊")
	}

	// FIXME: Add administrator validator
	switch cmd := strings.ToLower(commands[0]); cmd {
	case "clear":
		go func() {
			err := svc.cmdClearAllMsgReply(directChannel)
			if err != nil {
				svc.l.Errorf("execute command `%s` error, %+v", cmd, err)
			}
		}()
		return nil
	case "help":
		go func() {
			helperReply := "`clear` 清除所有機器人私訊" +
				"\n`info`  顯示機器人設定" +
				"\n`set`   更改機器人設定"
			err := svc.sendCommandReply(callbackUrl, helperReply)
			if err != nil {
				svc.l.Errorf("execute command `help` error, %+v", err)
			}
		}()
		return nil
	case "info":
		go func() {
			// TODO: Info command -> order of members and start time
		}()
		return nil
		// TODO: Set command -> send slack view to set settings
		// deal with
	default:
		return svc.sendCommandReply(callbackUrl, fmt.Sprintf("找不到指令行為 `%s`，執行 `help` 取得更多資訊", cmd))
	}
}

func (svc *SlackCommand) sendCommandReply(url util.Url, text string) error {
	notifier := util.NewSlackNotifier(svc.Token)
	msg := &util.GeneralMsg{
		Text: text + _commandClearNotification,
	}

	res, _, err := notifier.Send(svc.ctx, http.MethodPost, url, msg)
	if err != nil {
		fmt.Printf("bot send, %s\n", err)
	}

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
		_, _, err := notifier.Send(svc.ctx, http.MethodPost, util.Url("https://slack.com/api/chat.delete"), &util.GeneralMsg{
			Channel: directChannel,
			TS:      ts,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
