package service

import (
	"bitopi/intrernal/model"
	"bitopi/intrernal/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/logs"
)

const (
	_RootAdmin        = "<@U032TJB1PE1>"
	_CommandChannelID = "C01JS6YTHPE"
)

func (s *Service) MaidCommandHandler(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		fmt.Printf("parse form, %+v\n", errors.WithStack(err))
		return ok(c, "internal error")
	}

	payload := c.Request().PostForm
	callbackUrl := util.NewUrl(payload.Get("response_url"))
	userID := payload.Get("user_id")
	s.l.Debug("user: ", userID, ", from channel: ", payload.Get("channel_id"))
	directChannelID := s.getDirectChannelID(userID)
	if len(directChannelID) == 0 {
		return SendCommandReply(callbackUrl, "找不到私人訊息頻道")
	}
	fromChannelID := payload.Get("channel_id")
	s.l.Debugf("from channel ID: %s\nto channel ID: %s", fromChannelID, directChannelID)
	if directChannelID != fromChannelID {
		return SendCommandReply(callbackUrl, "指令只能在應用程式私訊使用，請到應用程式對話輸入指令")
	}

	caller := "<@" + userID + ">"
	valid, err := s.repo.IsAdmin(caller)
	if err != nil {
		return SendCommandReply(callbackUrl, fmt.Sprintf("無法驗證管理員, %s", err))
	}

	text := SplitText(payload.Get("text"))
	if len(text) == 0 {
		return SendCommandReply(callbackUrl, "需要輸入指令，執行 `/maid help` 取得更多資訊")
	}

	cmd := text[0]
	switch cmd {
	case "help":
		return SendCommandReply(callbackUrl, "*所有指令都只能在 <#"+_CommandChannelID+"> 使用*\n\n`/maid help` 顯示所有指令\n`/maid admin` 顯示所有管理員 ( 可執行指令人員 )\n`/maid admin {User}` 新增/刪除 管理員\n`/maid today` 回傳今日值班女僕\n`/maid list` 顯示女僕清單及起始計算日期\n`/maid set` 重設女僕清單及起始計算日期(7天換一次)\n```/maid set {UserA} {UserB} {UserC} ... {StartDate} \n/maid set @Yanun @Vic @Kai @Victor @Howard 2022-03-23```")
	case "list":
		maids := s.listMaid()
		t := s.getStartDate()
		return SendCommandReply(callbackUrl, "\n起算日期: "+t.Format("2006-01-02")+"\n女僕順序: "+strings.Join(maids, " "))
	case "set":
		if caller != _RootAdmin && !valid {
			return SendNoPermissionReply(s, callbackUrl)
		}
		users, message := parseContent(text[1:])
		t, err := time.Parse("2006-01-02", message[0])
		if err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("invalid time format, %s", err))
		}

		if err := s.repo.UpdateStartDate(t); err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("update time error, %s", err))
		}

		if err := s.repo.UpdateMaidList(users); err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("set maid error, %s", err))
		}
		channel := _CommandChannelID
		maids := s.listMaid()
		msg := caller + " 已重新設置女僕順序\n起算日期: " + s.getStartDate().Format("2006-01-02") + "\n女僕順序: " + strings.Join(maids, " ")
		sendChatPost(channel, msg)
		return SendCommandReply(callbackUrl, "set succeed")
	case "today":
		return SendCommandReply(callbackUrl, "今日女僕： "+s.getMaid())
	case "admin":
		if caller != _RootAdmin && !valid {
			return SendNoPermissionReply(s, callbackUrl)
		}
		if len(text) > 1 {
			users, _ := parseContent(text[1:])
			for i := range users {
				s.repo.ReverseAdmin(users[i])
			}
		}
		return SendCommandReply(callbackUrl, "*Admin list*\n"+s.getAdmin())
	default:
		return SendCommandReply(callbackUrl, fmt.Sprintf("找不到指令 `%s`，執行 `/maid help` 取得更多資訊", cmd))
	}
}

func (s *Service) getAdmin() string {
	admins, err := s.repo.ListAdmin()
	if err != nil {
		return _RootAdmin
	}
	res := strings.Join(admins, " ")
	if len(res) == 0 {
		res = _RootAdmin
	}
	return res
}

func sendChatPost(channel, text string) {
	bot := util.NewSlackNotifier(viper.GetString("token.maid"))
	res, code, err := bot.Send(context.Background(), util.PostChat, util.SlackMsg{
		Text:    text,
		Channel: channel,
	})
	if err != nil {
		fmt.Printf("bot send, %s\n", err)
	}
	l := logs.Get(context.Background())
	l.Debug("code: ", code)
	l.Debug("res: ", string(res))
}

func SendCommandReply(url util.Url, msg string) error {
	bot := util.NewSlackNotifier(viper.GetString("token.maid"))
	res, code, err := bot.Send(context.Background(), url, &util.GeneralMsg{Text: msg})
	if err != nil {
		fmt.Printf("bot send, %s\n", err)
	}
	l := logs.Get(context.Background())
	l.Debug("code: ", code)
	l.Debug("res: ", string(res))
	return nil
}

func SendNoPermissionReply(s *Service, callbackUrl util.Url) error {
	return SendCommandReply(callbackUrl, "用戶沒有執行此指令權限, 欲開啟權限請找管理員"+s.getAdmin())
}

func SplitText(text string) []string {
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

func parseContent(content []string) (users []string, message []string) {
	users = []string{}
	message = []string{}
	for i := range content {
		if content[i][0] == '<' && content[i][1] == '@' && content[i][len(content[i])-1] == '>' {
			users = append(users, content[i])
			continue
		}
		message = append(message, content[i])
	}
	return
}

func (s *Service) getDirectChannelID(userID string) string {
	url := "https://slack.com/api/conversations.open?users=" + userID
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		s.l.Errorf("create request error, %+v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+viper.GetString("token.maid")) //"Bearer "+viper.GetString("token.maid")
	s.l.Debugln(req.URL.String())
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		s.l.Errorf("send request error, %+v", err)
		return ""
	}

	buff, err := io.ReadAll(res.Body)
	if err != nil {
		s.l.Errorf("read response error, %+v", err)
		return ""
	}

	c := &model.DirectChannel{}
	if err := json.Unmarshal(buff, c); err != nil {
		s.l.Errorf("unmarshal json response error, %+v", err)
		return ""
	}
	if !c.OK {
		s.l.Errorf("invalid request, response:\n%s", string(buff))
		return ""
	}
	s.l.Debug("success get direct channel id " + c.Channel.ID)
	return c.Channel.ID
}
