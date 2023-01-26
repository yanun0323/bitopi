package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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
	Token               string
	DefaultStartDate    time.Time
	DefaultMemberList   []model.Member
	DefaultReplyMessage string
	DefaultMultiMember  bool
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
		return bot.ok(c)
	}

	if requestType == _eventVerification {
		return bot.ok(c, bot.verificationSlackResponse(c))
	}

	return bot.ok(c, bot.eventCallbackResponse(c))
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
		bot.l.Errorf("decode json error, %+v", err)
		return nil
	}
	bot.l.Debugf("slack event api: %+v", slackEventApi)

	exist, err := bot.recordMention(slackEventApi)
	if err != nil {
		bot.l.Errorf("record mention error, %+v", err)
		return nil
	}

	if exist {
		bot.l.Warnf("message was already replied, user: %s, channel: %s", slackEventApi.Event.User, slackEventApi.Event.Channel)
		return nil
	}

	dutyMember, leftMembers, err := bot.getDutyMember()
	if err != nil {
		bot.l.Errorf("get duty member error, %+v", err)
		return nil
	}

	rMsg, err := bot.getReplyMessage()
	if err != nil {
		return err
	}

	notifier := util.NewSlackNotifier(bot.Token)
	if err := bot.sendMentionReply(notifier, slackEventApi, dutyMember, leftMembers, rMsg); err != nil {
		bot.l.Errorf("send mention reply error, %+v", err)
		return nil
	}

	DMReceiver := []string{dutyMember}
	if rMsg.MultiMember {
		DMReceiver = append(DMReceiver, leftMembers...)
	}
	if err := bot.sendDirectMessage(notifier, slackEventApi, DMReceiver); err != nil {
		bot.l.Errorf("send direct message error, %+v", err)
		return nil
	}

	return nil
}

func (bot *SlackBot) recordMention(slackEventApi model.SlackEventAPI) (bool, error) {
	return bot.repo.FindOrCreateMentionRecord(bot.Name, slackEventApi.Event.Channel, slackEventApi.Event.EventTimeStamp)
}

func (bot *SlackBot) getDutyMember() (string, []string, error) {
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

	index := int(weekFromStartDate % int64(len(member)))
	bot.l.Debug("member list: ", member)
	bot.l.Debug("index: ", index)
	left := make([]string, 0, len(member)-1)
	for i, m := range member {
		if i != index {
			left = append(left, fmt.Sprintf("<@%s>", m))
		}
	}
	return fmt.Sprintf("<@%s>", member[index]), left, nil
}

func (bot *SlackBot) getReplyMessage() (model.SlackReplyMessage, error) {
	rMsg, err := bot.repo.GetReplyMessage(bot.Name)
	if err != nil {
		return model.SlackReplyMessage{}, err
	}

	if len(rMsg.Message) != 0 {
		return rMsg, nil
	}

	rMsg.Message = bot.DefaultReplyMessage
	rMsg.MultiMember = bot.DefaultMultiMember
	if err := bot.repo.SetReplyMessage(bot.Name, rMsg.Message, rMsg.MultiMember); err != nil {
		return model.SlackReplyMessage{}, err
	}
	return rMsg, nil
}

func (bot *SlackBot) getStartDate() time.Time {
	startDate, err := bot.repo.GetStartDate(bot.Name)
	if err != nil {
		bot.l.Warn("get start date error, %+v", err)
		bot.l.Warn("reset start date to database")
		bot.repo.UpdateStartDate(bot.Name, bot.DefaultStartDate)
		startDate = bot.DefaultStartDate
	}
	return startDate
}

func (bot *SlackBot) listMember() ([]string, error) {
	members, err := bot.repo.ListMember(bot.Name)
	if err == nil && len(members) != 0 {
		return members, nil
	}
	bot.l.Warnf("list member error, %+v", err)
	bot.l.Warnf("reset member to database '%s'", bot.Name)
	if err := bot.repo.UpdateMember(bot.Name, bot.DefaultMemberList); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(bot.DefaultMemberList))
	for _, m := range bot.DefaultMemberList {
		result = append(result, m.UserID)
	}

	return result, nil
}

func (bot *SlackBot) sendToSlack(notifier *util.SlackNotifier, msg util.Messenger) error {
	_, _, err := notifier.Send(bot.ctx, http.MethodPost, util.PostChat, msg)
	if err != nil {
		return err
	}

	return nil
}

func (bot *SlackBot) getPermalink(notifier *util.SlackNotifier, channel, messageTimestamp string) (string, error) {
	url := fmt.Sprintf("%s?channel=%s&message_ts=%s", util.GetPermalink, channel, messageTimestamp)
	response, _, err := notifier.Send(bot.ctx, http.MethodGet, util.Url(url), util.SlackMsg{})
	if err != nil {
		return "", err
	}

	permalink := model.SlackPermalinkResponse{}
	if err := json.Unmarshal(response, &permalink); err != nil {
		return "", err
	}

	if len(permalink.Error) != 0 {
		return "", errors.New(permalink.Error)
	}

	return permalink.Permalink, nil
}

func (bot *SlackBot) sendMentionReply(notifier *util.SlackNotifier, slackEventApi model.SlackEventAPI, dutyMember string, leftMembers []string, rMsg model.SlackReplyMessage) error {

	replyText := ""
	if rMsg.MultiMember {
		replyText = fmt.Sprintf(rMsg.Message, dutyMember, strings.Join(leftMembers, " "))
	} else {
		replyText = fmt.Sprintf(rMsg.Message, dutyMember)
	}

	if err := bot.sendToSlack(notifier, util.SlackReplyMsg{
		Text:      replyText,
		Channel:   slackEventApi.Event.Channel,
		TimeStamp: slackEventApi.Event.TimeStamp,
	}); err != nil {
		return err
	}
	return nil
}

func (bot *SlackBot) sendDirectMessage(notifier *util.SlackNotifier, slackEventApi model.SlackEventAPI, members []string) error {
	link, err := bot.getPermalink(notifier, slackEventApi.Event.Channel, slackEventApi.Event.EventTimeStamp)
	if err != nil {
		return err
	}

	directMessageText := fmt.Sprintf("*<%s|新訊息> 來自 <@%s> <#%s>*",
		link,
		slackEventApi.Event.User,
		slackEventApi.Event.Channel,
	)

	for _, member := range members {
		msg := util.SlackReplyMsg{
			Text:    directMessageText,
			Channel: member[2 : len(member)-1],
		}.AddAttachments(
			"text", slackEventApi.Event.Text,
			"callback_id", fmt.Sprintf("%s_direct_message_action", bot.Name),
			"actions", []model.SlackMessageButton{
				model.NewMessageActionButton("primary", "danger", "刪除通知"),
			})

		if err := bot.sendToSlack(notifier, msg); err != nil {
			return err
		}
	}
	return nil
}
