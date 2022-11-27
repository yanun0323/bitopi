package service

import (
	"bitopi/intrernal/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

const (
	DevopsDefaultStartTimeStr string = "20221023"
)

// 週數 | BitoEx | Meta |
// 本週 |   L    | T, H |
// 下週 |   T    | H, L |
// 下下 |   H    | L, T |

var (
	_devopsBroList = []string{
		"<@U03FDTNPWBW>", /* Lawrence */
		"<@U03RQKWLG8Z>", /* Tina */
		"<@U01A7LEG1CZ>", /* Harlan */
	}
)

func (s *Service) DevopsBotHandler(c echo.Context) error {
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

		bot := util.NewSlackNotifier(viper.GetString("token.devops"))

		exBroIndex := s.getExBroIndex()
		textPrefix := "請稍候片刻，本週猛哥/猛姐會盡快為您服務 :smiling_face_with_3_hearts:\n"
		textRow2 := fmt.Sprintf("Bito EX/Pro: %s\n", _devopsBroList[exBroIndex])
		textRow3 := fmt.Sprintf("Meta: %s", s.getMetaBroStr(exBroIndex))

		msg := util.SlackReplyMsg{
			Text:        textPrefix + textRow2 + textRow3,
			Channel:     slackEventApi.Event.Channel,
			UserName:    "Devops-Bro",
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

func (s *Service) getExBroIndex() int {

	start := s.getDevopsStartDate()
	now := time.Now()
	s.l.Debug("time start: ", start.Format("20060102 15:04:05 MST"))
	s.l.Debug("time now: ", now.Format("20060102 15:04:05 MST"))

	interval := now.Sub(start)
	week := int(((interval.Milliseconds()/1000/60)/60)/24) / 7
	s.l.Debug("week: ", week)

	index := week % len(_devopsBroList)
	return int(index)
}

func (s *Service) getDevopsStartDate() time.Time {
	t, _ := time.ParseInLocation("20060102", DevopsDefaultStartTimeStr, time.Local)
	return t
}

func (s *Service) getMetaBroStr(exBroIndex int) string {
	bros := make([]string, 0, len(_devopsBroList)-1)
	for i := range _devopsBroList {
		if i == exBroIndex {
			continue
		}
		bros = append(bros, _devopsBroList[i])
	}
	return strings.Join(bros, " ")
}
