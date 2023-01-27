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
	Name                string
	Token               string
	DefaultStartDate    time.Time
	DefaultMemberList   []model.Member
	DefaultReplyMessage string
	DefaultMultiMember  bool
}

func NewBot(svc Service, opt SlackBotOption) SlackBot {
	svc.l = logs.New(opt.Name, svc.logLevel)
	return SlackBot{
		Name:           opt.Name,
		Service:        svc,
		SlackBotOption: opt,
	}
}

func (svc *SlackBot) Handler(c echo.Context) error {
	requestType := svc.parseSlackRequestType(c)
	if len(requestType) == 0 {
		return svc.ok(c)
	}

	if requestType == _eventVerification {
		return svc.ok(c, svc.verificationSlackResponse(c))
	}

	return svc.ok(c, svc.eventCallbackResponse(c))
}

func (svc *SlackBot) parseSlackRequestType(c echo.Context) string {
	length, err := strconv.Atoi(c.Request().Header["Content-Length"][0])
	if err != nil {
		svc.l.Errorf("convert request header 'Content-Length' error, %+v", err)
		return ""
	}

	if length < 150 {
		return _eventVerification
	}

	return _eventCallback
}

func (svc *SlackBot) verificationSlackResponse(c echo.Context) interface{} {
	check := model.SlackTypeCheck{}
	if err := c.Bind(&check); err != nil {
		svc.l.Errorf("bind slack type check error, %+v", err)
		return nil
	}

	return model.SlackVerificationResponse{
		Challenge: check.Challenge,
	}
}

func (svc *SlackBot) eventCallbackResponse(c echo.Context) interface{} {
	slackEventApi := model.SlackEventAPI{}
	if err := json.NewDecoder(c.Request().Body).Decode(&slackEventApi); err != nil {
		svc.l.Errorf("decode json error, %+v", err)
		return nil
	}
	svc.l.Debugf("slack event api: %+v", slackEventApi)

	exist, err := svc.recordMention(slackEventApi)
	if err != nil {
		svc.l.Errorf("record mention error, %+v", err)
		return nil
	}

	if exist {
		svc.l.Warnf("message was already replied, user: %s, channel: %s", slackEventApi.Event.User, slackEventApi.Event.Channel)
		return nil
	}

	dutyMember, leftMembers, err := svc.getDutyMember()
	if err != nil {
		svc.l.Errorf("get duty member error, %+v", err)
		return nil
	}

	rMsg, err := svc.getReplyMessage()
	if err != nil {
		return err
	}

	notifier := util.NewSlackNotifier(svc.Token)
	go func() {
		err := svc.sendMentionReply(notifier, slackEventApi, dutyMember, leftMembers, rMsg)
		if err != nil {
			svc.l.Errorf("send mention reply error, %+v", err)
		}
	}()

	go func() {
		DMReceiver := []string{dutyMember}
		if rMsg.MultiMember {
			DMReceiver = append(DMReceiver, leftMembers...)
		}
		if err := svc.sendDirectMessage(notifier, slackEventApi, DMReceiver); err != nil {
			svc.l.Errorf("send direct message error, %+v", err)
		}
	}()

	return nil
}

func (svc *SlackBot) recordMention(slackEventApi model.SlackEventAPI) (bool, error) {
	return svc.repo.FindOrCreateMentionRecord(svc.Name, slackEventApi.Event.Channel, slackEventApi.Event.EventTimeStamp)
}

func (svc *SlackBot) getDutyMember() (string, []string, error) {
	startDate := svc.getStartDate()
	now := time.Now()
	svc.l.Debug("time start: ", startDate.Format("20060102 15:04:05 MST"))
	svc.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(startDate)
	weekFromStartDate := (((interval.Milliseconds() / 1000 / 60) / 60) / 24) / 7
	svc.l.Debug("week from start date: ", weekFromStartDate)

	member, err := svc.listMember()
	if err != nil {
		return "", nil, err
	}

	index := int(weekFromStartDate % int64(len(member)))
	svc.l.Debug("member list: ", member)
	svc.l.Debug("index: ", index)
	left := make([]string, 0, len(member)-1)
	for i, m := range member {
		if i != index {
			left = append(left, fmt.Sprintf("<@%s>", m))
		}
	}
	return fmt.Sprintf("<@%s>", member[index]), left, nil
}

func (svc *SlackBot) getReplyMessage() (model.SlackReplyMessage, error) {
	rMsg, err := svc.repo.GetReplyMessage(svc.Name)
	if err != nil {
		return model.SlackReplyMessage{}, err
	}

	if len(rMsg.Message) != 0 {
		return rMsg, nil
	}

	rMsg.Message = svc.DefaultReplyMessage
	rMsg.MultiMember = svc.DefaultMultiMember
	if err := svc.repo.SetReplyMessage(svc.Name, rMsg.Message, rMsg.MultiMember); err != nil {
		return model.SlackReplyMessage{}, err
	}
	return rMsg, nil
}

func (svc *SlackBot) getStartDate() time.Time {
	startDate, err := svc.repo.GetStartDate(svc.Name)
	if err != nil {
		svc.l.Warn("get start date error, %+v", err)
		svc.l.Warn("reset start date to database")
		svc.repo.UpdateStartDate(svc.Name, svc.DefaultStartDate)
		startDate = svc.DefaultStartDate
	}
	return startDate
}

func (svc *SlackBot) listMember() ([]string, error) {
	members, err := svc.repo.ListMember(svc.Name)
	if err == nil && len(members) != 0 {
		return members, nil
	}
	svc.l.Warnf("list member error, %+v", err)
	svc.l.Warnf("reset member to database '%s'", svc.Name)
	if err := svc.repo.UpdateMember(svc.Name, svc.DefaultMemberList); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(svc.DefaultMemberList))
	for _, m := range svc.DefaultMemberList {
		result = append(result, m.UserID)
	}

	return result, nil
}

func (svc *SlackBot) sendToSlack(notifier util.SlackNotifier, msg util.Messenger) error {
	_, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostChat, msg)
	if err != nil {
		return err
	}

	return nil
}

func (svc *SlackBot) getPermalink(notifier util.SlackNotifier, channel, messageTimestamp string) (string, error) {
	url := fmt.Sprintf("%s?channel=%s&message_ts=%s", util.GetPermalink, channel, messageTimestamp)
	response, _, err := notifier.Send(svc.ctx, http.MethodGet, util.Url(url), util.SlackMsg{})
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

func (svc *SlackBot) sendMentionReply(notifier util.SlackNotifier, slackEventApi model.SlackEventAPI, dutyMember string, leftMembers []string, rMsg model.SlackReplyMessage) error {
	replyText := ""
	if rMsg.MultiMember {
		replyText = fmt.Sprintf(rMsg.Message, dutyMember, strings.Join(leftMembers, " "))
	} else {
		replyText = fmt.Sprintf(rMsg.Message, dutyMember)
	}

	if err := svc.sendToSlack(notifier, util.SlackReplyMsg{
		Text:      replyText,
		Channel:   slackEventApi.Event.Channel,
		TimeStamp: slackEventApi.Event.TimeStamp,
	}); err != nil {
		return err
	}
	return nil
}

func (svc *SlackBot) sendDirectMessage(notifier util.SlackNotifier, slackEventApi model.SlackEventAPI, members []string) error {
	link, err := svc.getPermalink(notifier, slackEventApi.Event.Channel, slackEventApi.Event.EventTimeStamp)
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
			"text", "",
			"footer", slackEventApi.Event.Text+" ",
			"callback_id", fmt.Sprintf("%s_direct_message_action", svc.Name),
			"actions", []model.SlackMessageButton{
				model.NewMessageActionButton("primary", "danger", "刪除通知"),
			})

		if err := svc.sendToSlack(notifier, msg); err != nil {
			return err
		}
	}
	return nil
}
