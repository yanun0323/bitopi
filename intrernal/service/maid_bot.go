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
	MaidDefaultStartTimeStr string = "20220925"
)

var (
	_MaidList = []string{
		"<@U032TJB1PE1>", /* Yanun */
		"<@U03ECC8Q61E>", /* Howard */
		"<@U031SSN3QDT>", /* Kai */
		"<@U01QCKG7529>", /* Vic */
		"<@U036V8WPXDY>", /* Victor */
		"<@U03MWAJDBV3>", /* Luki */
	}
)

func (s *Service) MaidBotHandler(c echo.Context) error {
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
			s.l.Errorf("decode json, %s\n", err)
			return ok(c)
		}
		s.l.Debugf("%+v\n", slackEventApi)
		maid := s.getMaid()

		bot := util.NewSlackNotifier(viper.GetString("token.maid"))
		msg := util.SlackReplyMsg{
			Text:        fmt.Sprintf("請稍候片刻，本週女僕 %s 會盡快為您服務 :smiling_face_with_3_hearts:", maid),
			Channel:     slackEventApi.Event.Channel,
			UserName:    "Backend-Maid",
			TimeStamp:   slackEventApi.Event.TimeStamp,
			Attachments: make([]map[string]string, 0),
		}
		res, code, err := bot.Send(context.Background(), util.PostChat, msg)
		if err != nil {
			fmt.Printf("bot send, %s\n", err)
		}
		s.l.Debug("code: ", code)
		s.l.Debug("res: ", string(res))
		return ok(c)
	}

	return ok(c)
}

func (s *Service) getMaid() string {
	start := s.getStartDate()
	now := time.Now()
	s.l.Debug("time start: ", start.Format("20060102 15:04:05 MST"))
	s.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(start)
	week := (((interval.Milliseconds() / 1000 / 60) / 60) / 24) / 7
	s.l.Debug("week: ", week)

	maidList := s.listMaid()

	index := week % int64(len(maidList))
	return maidList[index]
}

func (s *Service) getStartDate() time.Time {
	t, err := s.repo.GetStartDate()
	if err != nil {
		s.l.Errorf("get start date error, %+v", err)
		s.l.Info("reset start date to database")
		t, _ = time.Parse("20060102", MaidDefaultStartTimeStr)
		s.repo.UpdateStartDate(t)
	}
	return t
}

func (s *Service) listMaid() []string {
	maidList, err := s.repo.ListMaid()
	if err != nil || len(maidList) == 0 {
		s.l.Errorf("list maid error, %+v", err)
		s.l.Info("reset maid list to database")
		_ = s.repo.UpdateMaidList(_MaidList)
		maidList = _MaidList
	}
	return maidList
}
