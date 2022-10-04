package service

import (
	"bitopi/intrernal/util"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	_RootAdmin = "<@U032TJB1PE1>"
)

func (svc *Service) MaidCommandHandler(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		fmt.Printf("parse form, %+v\n", errors.WithStack(err))
		return ok(c, "internal error")
	}

	payload := c.Request().PostForm
	callbackUrl := util.NewUrl(payload.Get("response_url"))
	caller := "<@" + payload.Get("user_id") + ">"
	valid, err := svc.repo.IsAdmin(caller)
	if err != nil {
		return SendCommandReply(callbackUrl, fmt.Sprintf("validate admin error, %s", err))
	}

	if caller != _RootAdmin && !valid {
		SendCommandReply(callbackUrl, "用戶沒有執行指令權限")
	}

	text := SplitText(payload.Get("text"))
	if len(text) == 0 {
		return SendCommandReply(callbackUrl, "需要輸入指令，執行 `/maid help` 取得更多資訊")
	}

	cmd := text[0]
	switch cmd {
	case "help":
		return SendCommandReply(callbackUrl, "*所有指令都只能在 <#C01JS6YTHPE> 使用*\n\n`/maid help` 顯示所有指令\n`/maid admin` 顯示所有管理員 ( 可執行指令人員 )\n`/maid admin {User}` 新增/刪除 管理員\n`/maid today` 回傳今日值班女僕\n`/maid list` 顯示女僕清單及起始計算日期\n`/maid set` 重設女僕清單及起始計算日期(7天換一次)\n```/maid set {UserA} {UserB} {UserC} ... {StartDate} \n/maid set @Yanun @Vic @Kai @Victor @Howard 2022-03-23```")
	case "list":
		maids := svc.listMaid()
		t := svc.getStartDate()
		return SendCommandReply(callbackUrl, strings.Join(maids, " ")+" "+t.Format("2006-01-02"))
	case "set":
		users, message := parseContent(text[1:])
		t, err := time.Parse("2006-01-02", message[0])
		if err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("invalid time format, %s", err))
		}

		if err := svc.repo.UpdateStartDate(t); err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("update time error, %s", err))
		}

		if err := svc.repo.UpdateMaidList(users); err != nil {
			return SendCommandReply(callbackUrl, fmt.Sprintf("set maid error, %s", err))
		}

		channel := payload.Get("channel_id")
		maids := svc.listMaid()
		msg := caller + " 已重新設置女僕順序\n" + strings.Join(maids, " ") + " " + svc.getStartDate().Format("2006-01-02")
		sendChatPost(channel, msg)
		return SendCommandReply(callbackUrl, "set succeed")
	case "today":
		return SendCommandReply(callbackUrl, "今日女僕： "+svc.getMaid())
	case "admin":
		if len(text) > 1 {
			users, _ := parseContent(text[1:])
			for i := range users {
				svc.repo.ReverseAdmin(users[i])
			}
		}
		admins, err := svc.repo.ListAdmin()
		if err != nil {
			return ok(c, err)
		}
		res := strings.Join(admins, " ")
		if len(res) == 0 {
			res = "empty"
		}
		return SendCommandReply(callbackUrl, "*Admin list*\n"+res)
	default:
		return SendCommandReply(callbackUrl, fmt.Sprintf("找不到指令 `%s`，執行 `/maid help` 取得更多資訊", cmd))
	}
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
	fmt.Println("code: ", code)
	fmt.Println("res: ", string(res))
}

func SendCommandReply(url util.Url, msg string) error {
	bot := util.NewSlackNotifier(viper.GetString("token.maid"))
	res, code, err := bot.Send(context.Background(), url, &util.GeneralMsg{Text: msg})
	if err != nil {
		fmt.Printf("bot send, %s\n", err)
	}
	fmt.Println("code: ", code)
	fmt.Println("res: ", string(res))
	return nil
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
