package service

import (
	"bitopi/intrernal/util"
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
	PMDefaultStartTimeStr string = "20221127"
)

var (
	_pmList = []string{
		"<@U02223HG26L>", /* Rafeni */
		"<@U01THK4U2MD>", /* Momo */
		"<@UGVSBTC94>",   /* Donii */
	}
	_pmNameList = []string{
		"Rafeni", /* Rafeni */
		"Momo",   /* Momo */
		"Donii",  /* Donii */
	}
)

func (s *Service) PMBotHandler(c echo.Context) error {
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

		bot := util.NewSlackNotifier(viper.GetString("token.pm"))

		pmIndex, weekend := s.getPMIndex()
		w := "本週"
		if weekend {
			w = "週末"
		}
		replyText := fmt.Sprintf("請稍候片刻，%s Support PM %s 將盡快為您服務 :smiling_face_with_3_hearts:\n", w, _pmList[pmIndex])

		msg := util.SlackReplyMsg{
			Text:        replyText,
			Channel:     slackEventApi.Event.Channel,
			UserName:    "PM",
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

func (s *Service) getPMIndex() (int, bool) {
	start := s.getPMStartDate()
	now := time.Now()
	s.l.Debug("time start: ", start.Format("20060102 15:04:05 MST"))
	s.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	if isWeekend(now) { /* 假日Don */
		return 2, true
	}

	interval := now.Sub(start)
	week := int(((interval.Milliseconds()/1000/60)/60)/24) / 7
	s.l.Debug("week: ", week)

	index := week % len(_devopsBroList)
	return int(index), false
}

func (s *Service) getPMStartDate() time.Time {
	t, _ := time.ParseInLocation("20060102", PMDefaultStartTimeStr, time.Local)
	return t
}

func isWeekend(t time.Time) bool {
	w := t.Weekday()
	return w == time.Saturday || w == time.Sunday
}
