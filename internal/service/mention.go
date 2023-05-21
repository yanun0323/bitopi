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
	"github.com/pkg/errors"
	"github.com/yanun0323/pkg/logs"
)

const (
	_eventVerification = "url_verification"
	_eventCallback     = "event_callback"
)

type SlackBot struct {
	Service
	SlackBotOption
}

type SlackBotOption struct {
	Name                      string
	Token                     string
	DefaultStartDate          time.Time
	DefaultDutyDuration       time.Duration
	DefaultMemberCountPerTime int
	DefaultMemberList         []model.Member
	DefaultReplyMessage       string
	DefaultHomeReplyMessage   string
	DefaultMultiMember        bool
}

func NewBot(svc Service, opt SlackBotOption) SlackBot {
	svc.l = logs.New(opt.Name, svc.logLevel)
	return SlackBot{
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
		svc.l.Errorf("convert request header 'Content-Length' failed, err: %+v", err)
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
		svc.l.Errorf("bind slack type check failed, err: %+v", err)
		return nil
	}

	return model.SlackVerificationResponse{
		Challenge: check.Challenge,
	}
}

func (svc *SlackBot) eventCallbackResponse(c echo.Context) interface{} {
	slackEventApi := model.SlackEventAPI{}
	if err := json.NewDecoder(c.Request().Body).Decode(&slackEventApi); err != nil {
		svc.l.Errorf("decode json failed, err: %+v", err)
		return nil
	}
	svc.l.Debugf("slack event api: %+v", slackEventApi)

	id, exist, err := svc.recordMention(slackEventApi)
	if err != nil {
		svc.l.Errorf("record mention failed, err: %+v", err)
		return nil
	}

	if exist {
		svc.l.Warnf("message was already replied, user: %s, channel: %s", slackEventApi.Event.User, slackEventApi.Event.Channel)
		return nil
	}

	startDate := svc.getStartDate()
	dutyDuration := svc.getDutyDuration()
	dutyMemberCountPerTime := svc.getDutyMemberCountPerTime()
	dutyMember, leftMembers, err := svc.getDutyMember(true, startDate, dutyMemberCountPerTime, dutyDuration)
	if err != nil {
		svc.l.Errorf("get duty member failed, err: %+v", err)
		return nil
	}

	rMsg, err := svc.getReplyMessage()
	if err != nil {
		return err
	}

	notifier := util.NewSlackNotifier(svc.Token)
	res, err := svc.sendMentionReply(notifier, slackEventApi, dutyMember, leftMembers, rMsg)
	if err != nil {
		svc.l.Errorf("send mention reply failed, err: %+v", err)
	}
	resMap, err := util.ParseByte2Map(res)
	if err != nil {
		svc.l.Errorf("parse slack response failed, err: %+v", err)
		return err
	}
	resChannel, ok := resMap["channel"].(string)
	if !ok {
		err := errors.New("get channel from slack response failed")
		svc.l.Error(err)
		return err
	}
	resTS, ok := resMap["ts"].(string)
	if !ok {
		err := errors.New("get ts from slack response failed")
		svc.l.Error(err)
		return err
	}

	go func() {
		receiveMembers := dutyMember
		if rMsg.MentionMultiMember {
			receiveMembers = append(receiveMembers, leftMembers...)
		}
		if err := svc.sendReplyDirectMessage(notifier, model.SlackDirectMsgOption{
			IsUser:          true,
			MentionRecordID: fmt.Sprintf("%d", id),
			ServiceName:     svc.Name,
			User:            slackEventApi.Event.User,
			EventContent:    slackEventApi.Event.Text,
			Members:         receiveMembers,
			LinkChannel:     resChannel,
			LinkTimestamp:   resTS,
		}); err != nil {
			svc.l.Errorf("send direct message failed, err: %+v", err)
			return
		}

		if err := svc.publishHomeView(notifier); err != nil {
			svc.l.Errorf("publish home view failed, err: %+v", err)
		}
	}()

	return nil
}

func (svc *SlackBot) recordMention(slackEventApi model.SlackEventAPI) (uint64, bool, error) {
	return svc.repo.FindOrCreateMentionRecord(svc.Name, slackEventApi.Event.Channel, slackEventApi.Event.EventTimeStamp)
}

func (svc *SlackBot) getDutyMember(mention bool, startDate time.Time, dutyCount int, dutyDuration time.Duration) ([]string, []string, error) {
	now := time.Now()
	svc.l.Debug("time start: ", startDate.Format("20060102 15:04:05 MST"))
	svc.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(startDate)
	weekFromStartDate := (((interval.Milliseconds() / 1000 / 60) / 60) / 24) / 7
	svc.l.Debug("week from start date: ", weekFromStartDate)

	member, err := svc.listMember(mention)
	if err != nil {
		return nil, nil, err
	}

	dutyWeek := int(dutyDuration / (time.Hour * 24 * 7))
	passedRound := (int(weekFromStartDate) / dutyWeek)
	index := passedRound * dutyCount % len(member)

	svc.l.Debug("member list: ", member)
	svc.l.Debug("index: ", index)
	duty := make([]string, 0, dutyCount)
	left := make([]string, 0, len(member)-dutyCount)
	for ; len(duty) < dutyCount; index++ {
		if index >= len(member) {
			index = 0
		}
		duty = append(duty, member[index])
		member[index] = ""
	}

	for _, mem := range member {
		if len(mem) == 0 {
			continue
		}
		left = append(left, mem)
	}

	return duty, left, nil
}

func (svc *SlackBot) getReplyMessage() (model.BotMessage, error) {
	msg, err := svc.repo.GetReplyMessage(svc.Name)
	if err != nil {
		return model.BotMessage{}, err
	}

	if msg.ID != 0 {
		return msg, nil
	}

	msg.Service = svc.Name
	msg.MentionMultiMember = svc.DefaultMultiMember
	msg.MentionMessage = svc.DefaultReplyMessage
	msg.HomeMentionMessage = svc.DefaultHomeReplyMessage
	if len(msg.HomeMentionMessage) == 0 {
		msg.HomeMentionMessage = msg.MentionMessage
	}

	if err := svc.repo.SetReplyMessage(msg); err != nil {
		return model.BotMessage{}, err
	}
	return msg, nil
}

func (svc *SlackBot) getStartDate() time.Time {
	startDate, err := svc.repo.GetStartDate(svc.Name)
	if err != nil {
		svc.l.Warn("get start date failed, err: %+v", err)
		svc.l.Warn("reset start date to database")
		svc.repo.UpdateStartDate(svc.Name, svc.DefaultStartDate)
		startDate = svc.DefaultStartDate
	}
	return startDate
}

func (svc *SlackBot) getDutyDuration() time.Duration {
	dutyDuration, err := svc.repo.GetDutyDuration(svc.Name)
	if err != nil || dutyDuration == 0 {
		svc.l.Warnf("get duty duration, err: %+v", err)
		return svc.DefaultDutyDuration
	}
	return dutyDuration
}

func (svc *SlackBot) getDutyMemberCountPerTime() int {
	dutyMemberCountPerTime, err := svc.repo.GetDutyMemberCountPerTime(svc.Name)
	if err != nil || dutyMemberCountPerTime == 0 {
		svc.l.Warnf("get duty member count per time, err: %+v", err)
		return svc.DefaultMemberCountPerTime
	}
	return dutyMemberCountPerTime
}

func (svc *SlackBot) listMember(mention bool) ([]string, error) {
	members, err := svc.repo.ListMembers(svc.Name)
	if err == nil && len(members) != 0 {
		return svc.transferMembersToString(members, mention), nil
	}
	svc.l.Warnf("list member failed, err: %+v", err)
	svc.l.Warnf("reset member to database '%s'", svc.Name)
	if err := svc.repo.ResetMembers(svc.Name, svc.DefaultMemberList); err != nil {
		return nil, err
	}

	return svc.transferMembersToString(svc.DefaultMemberList, mention), nil
}

func (svc *SlackBot) transferMembersToString(members []model.Member, mention bool) []string {
	s := make([]string, 0, len(members))
	for _, member := range members {
		if mention {
			s = append(s, member.UserTag())
			continue
		}
		s = append(s, member.UserID)
	}
	return s
}

func (svc *SlackBot) sendMentionReply(notifier util.SlackNotifier, slackEventApi model.SlackEventAPI, dutyMember []string, leftMembers []string, rMsg model.BotMessage) ([]byte, error) {
	replyText := ""
	if rMsg.MentionMultiMember {
		replyText = fmt.Sprintf(rMsg.MentionMessage, strings.Join(dutyMember, " "), strings.Join(leftMembers, " "))
	} else {
		replyText = fmt.Sprintf(rMsg.MentionMessage, strings.Join(dutyMember, " "))
	}

	res, err := svc.postMessage(notifier, util.SlackReplyMsg{
		Text:      replyText,
		Channel:   slackEventApi.Event.Channel,
		TimeStamp: slackEventApi.Event.TimeStamp,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
