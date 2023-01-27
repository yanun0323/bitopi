package service

import (
	"bitopi/internal/util"
	"bitopi/pkg/infra"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/yanun0323/pkg/logs"
)

func TestDeleteBotDirectMessage(t *testing.T) {
	mustNil := func(i interface{}) {
		require.Nil(t, i)
	}
	mustNil(infra.Init(""))
	l := logs.New("maid", 2)
	notifier := util.NewSlackNotifier(viper.GetString("maid.token"))
	ctx := context.Background()
	res, _, err := notifier.Send(ctx, http.MethodGet, util.Url("https://slack.com/api/conversations.history?channel=D03C5N7U2LA"), &util.GeneralMsg{})
	mustNil(err)

	data := map[string]interface{}{}
	mustNil(json.Unmarshal(res, &data))
	for k, v := range data {
		l.Info(k, ": ", v)
	}

	if data["messages"] == nil {
		l.Error("empty messages")
		return
	}

	for _, msg := range data["messages"].([]interface{}) {
		ts := msg.(map[string]interface{})["ts"].(string)
		res, _, err := notifier.Send(ctx, http.MethodPost, util.Url("https://slack.com/api/chat.delete"), &util.SlackSimpleMsg{
			Channel:   "D03C5N7U2LA",
			Timestamp: ts,
		})

		if err != nil {
			l.Errorf("%+v", err)
			return
		}

		l.Debugf("%s", string(res))
	}
}

type TestMsg struct {
	Channel string `json:"channel"`
}

func (msg *TestMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}
