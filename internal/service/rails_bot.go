package service

import (
	"bitopi/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

const (
	RailsDefaultStartTimeStr string = "20221106"
)

var (
	_railsHiList = []string{
		"<@U0156SRG9GW>", /* Gmi */
		"<@UKL1DAL4E>",   /* Barry */
		"<@U0328D2JE8H>", /* Kurt */
		"<@UQTAPAF2T>",   /* Kevin */
		"<@U041HD3AQ3D>", /* Eric */
		"<@U01GTQ8K52P>", /* Yuan */
	}
)

func (s *Service) RailsBotHandler(c echo.Context) error {
	length, err := strconv.Atoi(c.Request().Header["Content-Length"][0])
	if err != nil {
		return ok(c)
	}

	v := "event_callback"
	if length < 150 {
		v = "url_verification"
	}

	s.l.Debug("type: ", v)
	switch v {
	case "url_verification":
		v := &util.SlackTypeCheck{}
		if err := c.Bind(&v); err != nil {
			return ok(c)
		}
		return ok(c, util.SlackVerificationResponse{
			Challenge: v.Challenge,
		})
	case "event_callback":
		slackEventApi := util.SlackEventAPI{}
		err := json.NewDecoder(c.Request().Body).Decode(&slackEventApi)
		if err != nil && err != io.EOF {
			s.l.Debugf("decode json, %s\n", err)
			return ok(c)
		}
		s.l.Debugf("%+v\n", slackEventApi)

		bot := util.NewSlackNotifier(viper.GetString("token.rails"))

		railsIndex := s.getRailsIndex()
		replyText := fmt.Sprintf("請稍候片刻，本週茅房廁紙 %s 會盡快為您服務 :smiling_face_with_3_hearts:\n", _railsHiList[railsIndex])

		msg := util.SlackReplyMsg{
			Text:        replyText,
			Channel:     slackEventApi.Event.Channel,
			UserName:    "Rails-Hi",
			TimeStamp:   slackEventApi.Event.TimeStamp,
			Attachments: make([]map[string]string, 0),
		}
		res, code, err := bot.Send(context.Background(), util.PostChat, msg)
		if err != nil {
			s.l.Debugf("bot send, %s\n", err)
		}
		s.l.Debug("code: ", code)
		s.l.Debug("res: ", string(res))
		return ok(c)
	}

	return ok(c)
}

func (s *Service) getRailsIndex() int {

	start := s.getRailsStartDate()
	now := time.Now()
	s.l.Debug("time start: ", start.Format("20060102 15:04:05 MST"))
	s.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(start)
	week := int(((interval.Milliseconds()/1000/60)/60)/24) / 7
	s.l.Debug("week: ", week)

	index := week % len(_railsHiList)
	return int(index)
}

func (s *Service) getRailsStartDate() time.Time {
	t, _ := time.ParseInLocation("20060102", RailsDefaultStartTimeStr, time.Local)
	return t
}
