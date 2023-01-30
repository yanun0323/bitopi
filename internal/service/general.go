package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (svc *Service) ok(c echo.Context, i ...interface{}) error {
	svc.l.Info("OK")
	if len(i) > 0 {
		return c.JSON(http.StatusOK, i[0])
	}
	return c.JSON(http.StatusOK, nil)
}

func (svc *Service) postMessage(notifier util.SlackNotifier, msg util.Messenger) error {
	res, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostChat, msg)
	if err != nil {
		return err
	}
	svc.l.Debugf("response:\n%s", string(res))
	return nil
}

func (svc *Service) getMessage(notifier util.SlackNotifier, channel, ts string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?channel=%s&ts=%s", util.GetChat, channel, ts)
	res, _, err := notifier.Send(svc.ctx, http.MethodGet, util.Url(url), &util.GeneralMsg{})
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal(res, &m); err != nil {
		return nil, err
	}

	if m["messages"] == nil {
		return nil, errors.New("nil messages")
	}

	msgs := m["messages"].([]interface{})
	if len(msgs) == 0 {
		return nil, errors.New("empty messages")
	}

	for _, raw := range msgs {
		msg := raw.(map[string]interface{})
		if msg["ts"] == ts {
			return msg, nil
		}
	}

	return nil, errors.New("can't find match message")
}

func (svc *Service) getDirectChannel(userID, token string) (string, error) {
	buf, _, err := util.HttpRequest(util.HttpRequestOption{
		Method:       http.MethodPost,
		Url:          "https://slack.com/api/conversations.open?users=" + userID,
		Token:        token,
		IsUrlencoded: true,
	})
	if err != nil {
		return "", err
	}

	c := &model.SlackDirectChannel{}
	if err := json.Unmarshal(buf, c); err != nil {
		return "", err
	}

	if !c.OK {
		return "", errors.Errorf("invalid request, response:\n%s", string(buf))
	}

	svc.l.Debugf("success get direct channel id %s", c.Channel.ID)
	return c.Channel.ID, nil
}

func (svc *Service) getPermalink(notifier util.SlackNotifier, channel, messageTimestamp string) (string, error) {
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

func (svc *Service) sendReplyDirectMessage(notifier util.SlackNotifier, opt model.SlackDirectMsgOption) error {
	link, err := svc.getPermalink(notifier, opt.Channel, opt.EventTimestamp)
	if err != nil {
		return err
	}

	directMessageText := fmt.Sprintf("*<%s|新訊息> 來自 <@%s> <#%s>*",
		link,
		opt.User,
		opt.Channel,
	)

	for _, member := range opt.Members {
		ch := member[2 : len(member)-1]
		if !opt.IsUser {
			ch, err = svc.getDirectChannel(member, notifier.Token())
			if err != nil {
				return err
			}
		}
		msg := util.SlackReplyMsg{
			Text:    directMessageText,
			Channel: ch,
		}.AddAttachments(
			"type", "section",
			"text", "",
			"footer", opt.EventContent,
			"callback_id", fmt.Sprintf("%s_direct_message_action", opt.ServiceName),
			"actions", []model.SlackActionButton{
				model.NewSlackActionButton("primary", opt.MentionRecordID, "轉傳給..."),
				model.NewSlackActionButton("danger", "delete", "刪除"),
			},
		)

		if err := svc.postMessage(notifier, msg); err != nil {
			return err
		}
	}
	return nil
}
