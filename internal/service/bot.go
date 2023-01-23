package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yanun0323/pkg/logs"
)

const (
	_eventVerification = "url_verification"
	_eventCallback     = "event_callback"
)

type SlackBot struct {
	Name string
	Service
	SlackBotOption
}

type SlackBotOption struct {
	DefaultStartDate  time.Time
	DefaultMemberList []string
	MemberTableName   string
	Token             string
	ReplyMsgFormat    string
	IsMultiMember     bool
}

func NewBot(name string, svc Service, opt SlackBotOption) SlackBot {
	svc.l = logs.New(name, svc.logLevel)
	return SlackBot{
		Name:           name,
		Service:        svc,
		SlackBotOption: opt,
	}
}

func (bot *SlackBot) Handler(c echo.Context) error {
	requestType := bot.parseSlackRequestType(c)
	if len(requestType) == 0 {
		return ok(c)
	}

	if requestType == _eventVerification {
		return ok(c, bot.verificationSlackResponse(c))
	}

	return ok(c, bot.eventCallbackResponse(c))
}

func (bot *SlackBot) parseSlackRequestType(c echo.Context) string {
	length, err := strconv.Atoi(c.Request().Header["Content-Length"][0])
	if err != nil {
		bot.l.Errorf("convert request header 'Content-Length' error, %+v", err)
		return ""
	}

	if length < 150 {
		return _eventVerification
	}

	return _eventCallback
}

func (bot *SlackBot) verificationSlackResponse(c echo.Context) interface{} {
	check := model.SlackTypeCheck{}
	if err := c.Bind(&check); err != nil {
		bot.l.Errorf("bind slack type check error, %+v", err)
		return nil
	}

	return model.SlackVerificationResponse{
		Challenge: check.Challenge,
	}
}

func (bot *SlackBot) eventCallbackResponse(c echo.Context) interface{} {
	slackEventApi := model.SlackEventAPI{}
	if err := json.NewDecoder(c.Request().Body).Decode(&slackEventApi); err != nil {
		bot.l.Errorf("decode json, %+v", err)
		return nil
	}

	bot.l.Debugf("slack event api: %+v", slackEventApi)
	dutyMember, leftMembers, err := bot.GetDutyMember()
	if err != nil {
		bot.l.Errorf("get duty member error, %+v", err)
		return nil
	}
	text := ""
	if bot.IsMultiMember {
		text = fmt.Sprintf(bot.ReplyMsgFormat, dutyMember, strings.Join(leftMembers, " "))
	} else {
		text = fmt.Sprintf(bot.ReplyMsgFormat, dutyMember)
	}

	notifier := util.NewSlackNotifier(bot.Token)
	msg := util.SlackReplyMsg{
		Text:        text,
		Channel:     slackEventApi.Event.Channel,
		TimeStamp:   slackEventApi.Event.TimeStamp,
		Attachments: []map[string]string{},
	}

	response, code, err := notifier.Send(bot.ctx, util.PostChat, msg)
	if err != nil {
		bot.l.Warnf("send message to slack error, %+v", err)
		return ok(c)
	}

	bot.l.Debugf("code: %d, response: %s", code, string(response))
	return ok(c)
}

func (bot *SlackBot) GetDutyMember() (string, []string, error) {
	startDate := bot.getStartDate()
	now := time.Now()
	bot.l.Debug("time start: ", startDate.Format("20060102 15:04:05 MST"))
	bot.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(startDate)
	weekFromStartDate := (((interval.Milliseconds() / 1000 / 60) / 60) / 24) / 7
	bot.l.Debug("week from start date: ", weekFromStartDate)

	member, err := bot.listMember()
	if err != nil {
		return "", nil, err
	}

	index := weekFromStartDate % int64(len(member))
	return member[index], append(member[:index], member[index+1:]...), nil
}

func (bot *SlackBot) getStartDate() time.Time {
	// TODO: Add service start date key in database
	// bot.repo.GetStartDate(bot.MemberTableName)
	// Table be like: | service_name | date |
	startDate, err := bot.repo.GetStartDate()
	if err != nil {
		bot.l.Warn("get start date error, %+v", err)
		bot.l.Warn("reset start date to database")
		bot.repo.UpdateStartDate(bot.DefaultStartDate)
		startDate = bot.DefaultStartDate
	}
	return startDate
}

func (bot *SlackBot) listMember() ([]string, error) {
	member, err := bot.repo.ListMember(bot.MemberTableName)
	if err == nil && len(member) != 0 {
		return member, nil
	}
	bot.l.Warnf("list member error, %+v", err)
	bot.l.Warnf("reset %s member to database '%s'", bot.MemberTableName)
	if err := bot.repo.UpdateMember(bot.MemberTableName, bot.DefaultMemberList); err != nil {
		return nil, err
	}
	return bot.DefaultMemberList, nil
}
