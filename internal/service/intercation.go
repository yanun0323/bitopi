package service

import (
	"bitopi/internal/model"
	"bitopi/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type SlackInteraction struct {
	SlackBot
}

type SlackInteractionOption struct {
}

func NewInteraction(bot SlackBot) SlackInteraction {
	return SlackInteraction{
		SlackBot: bot,
	}
}

func (svc *SlackInteraction) Handler(c echo.Context) error {
	return svc.ok(c, svc.interactionResponse(c))
}

func (svc *SlackInteraction) interactionResponse(c echo.Context) interface{} {
	payload, err := svc.parsePayload(c)
	if err != nil || payload == nil {
		svc.l.Errorf("parse payload failed, err: %+v", err)
	}

	switch payload["type"].(string) {
	case "view_submission": /* .handle action from action block view */
		return svc.viewSubmissionHandler(c, payload)
	case "interactive_message": /* handle action from button of bot direct message */
		return svc.actionHandler(payload)
	case "block_actions": /* handle action from button of bot home*/
		return svc.homeHandler(payload)
	default:
		return svc.noneInteractionReply(payload)
	}
}

func (svc *SlackInteraction) parsePayload(c echo.Context) (map[string]interface{}, error) {
	data := map[string]string{}
	if err := c.Bind(&data); err != nil {
		return nil, err
	}

	if len(data["payload"]) == 0 {
		return nil, errors.New("empty payload")
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data["payload"]), &payload); err != nil {
		return nil, err
	}

	for k, v := range payload {
		if k == "original_message" {
			svc.l.Debug("--- original_message ---")
			for kk, vv := range v.(map[string]interface{}) {
				svc.l.Debug(kk, ": ", vv)
			}
			svc.l.Debug("------")
			continue
		}

		svc.l.Debug(k, ": ", v)
	}

	return payload, nil
}

func (svc *SlackInteraction) parseActionValue(payload map[string]interface{}) (id, value string, err error) {
	v, err := svc.parseAction(payload)
	if err != nil {
		return "", "", err
	}

	vs := strings.Split(v, ",")
	if len(vs) != 2 {
		return "", "", errors.Errorf("mismatch action value length: %d", len(vs))
	}
	return vs[0], vs[1], nil
}

func (svc *SlackInteraction) actionHandler(payload map[string]interface{}) interface{} {
	id, action, err := svc.parseActionValue(payload)
	if err != nil {
		svc.l.Errorf("parse action, err: %+v", err)
		return svc.noneInteractionReply(payload)
	}

	svc.l.Infof("action: %s", action)
	switch action {
	case "delete":
		return svc.deleteActionReply()
	case "delete.and.reply":
		return svc.deleteAndReplyActionReply(id)
	case "resend":
		return svc.resendActionReply(id, payload)
	default:
		svc.l.Warnf("unknown action: %s", action)
		return svc.noneInteractionReply(payload)
	}
}

func (svc *SlackInteraction) parseAction(payload map[string]interface{}) (string, error) {
	if payload["actions"] == nil {
		return "", errors.New("empty actions content")
	}

	actions, ok := payload["actions"].([]interface{})
	if !ok {
		return "", errors.New("transfer actions type error")
	}

	value, exist := actions[0].(map[string]interface{})["value"]
	if !exist {
		return "", errors.New("empty actions value")
	}

	svc.l.Debug("action value: ", value)
	return value.(string), nil
}

func (svc *SlackInteraction) resendActionReply(mentionID string, payload map[string]interface{}) interface{} {
	svc.l.Debug("execute resend")
	go func() {
		notifier := util.NewSlackNotifier(svc.Token)
		_, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostView,
			resendView(payload["trigger_id"].(string), mentionID),
		)
		if err != nil {
			svc.l.Errorf("send resend action view, err: %+v", err)
			return
		}
	}()

	return svc.noneInteractionReply(payload)
}

func (svc *SlackInteraction) deleteActionReply() interface{} {
	svc.l.Debug("execute delete")
	return struct {
		DeleteOriginal bool `json:"delete_original"`
	}{
		DeleteOriginal: true,
	}
}

func (svc *SlackInteraction) deleteAndReplyActionReply(mentionID string) interface{} {
	svc.l.Debug("execute delete and reply")
	id, err := strconv.Atoi(mentionID)
	if err != nil {
		svc.l.Errorf("convert mention ID, err: %+v", err)
		return nil
	}

	record, err := svc.repo.GetMentionRecord(svc.ctx, uint64(id))
	if err != nil {
		svc.l.Errorf("get mention record, err: %+v", err)
		return nil
	}

	text := "已處理，有需要再 Tag 我 ☺️"
	msg, err := svc.getReplyMessage()
	if err == nil && len(msg.DoneReplyMessage) != 0 {
		text = msg.DoneReplyMessage
	}

	notifier := util.NewSlackNotifier(svc.Token)
	svc.postMessage(notifier, util.SlackReplyMsg{
		Text:      text,
		Channel:   record.Channel,
		TimeStamp: record.Timestamp,
	})

	return struct {
		DeleteOriginal bool `json:"delete_original"`
	}{
		DeleteOriginal: true,
	}
}

func (svc *SlackInteraction) noneInteractionReply(payload map[string]interface{}) interface{} {
	svc.l.Debug("execute original")
	return payload["original_message"]
}

func resendView(triggerID, data string) util.SlackViewMsg {
	return util.SlackViewMsg{
		TriggerID: triggerID,
		View: map[string]interface{}{
			"private_metadata": data,
			"type":             "modal",
			"submit": util.PlainText{
				Type:  "plain_text",
				Text:  "轉傳",
				Emoji: true,
			},
			"close": util.PlainText{
				Type:  "plain_text",
				Text:  "取消",
				Emoji: true,
			},
			"title": util.PlainText{
				Type:  "plain_text",
				Text:  "轉傳給...",
				Emoji: true,
			},
			"blocks": `[
				{
					"type": "input",
					"element": {
						"type": "multi_users_select",
						"placeholder": {
							"type": "plain_text",
							"text": "選擇人員",
							"emoji": true
						},
						"action_id": "multi_users_select-action"
					},
					"label": {
						"type": "plain_text",
						"text": "選擇要轉傳通知的使用者(可複選)",
						"emoji": true
					}
				},
				{
					"type": "context",
					"elements": [
						{
							"type": "plain_text",
							"text": "＃通知將會透過機器人轉傳",
							"emoji": true
						}
					]
				}
			]`,
		},
	}
}

// TODO: Add resend user to resend message
func (svc *SlackInteraction) viewSubmissionHandler(c echo.Context, payload map[string]interface{}) interface{} {
	svc.l.Debug("handle view submission")
	go func() {
		view := payload["view"].(map[string]interface{})
		mentionID := view["private_metadata"].(string)
		id, err := strconv.Atoi(mentionID)
		if err != nil {
			svc.l.Errorf("convert mention ID, err: %+v", err)
			return
		}

		record, err := svc.repo.GetMentionRecord(svc.ctx, uint64(id))
		if err != nil {
			svc.l.Errorf("get mention record, err: %+v", err)
			return
		}

		notifier := util.NewSlackNotifier(svc.Token)
		msg, err := svc.getMessage(notifier, record.Channel, record.Timestamp)
		if err != nil {
			svc.l.Errorf("get message from slack, err: %+v", err)
			return
		}

		values := view["state"].(map[string]interface{})["values"].(map[string]interface{})
		resendUserID := payload["user"].(map[string]interface{})["id"].(string)
		users := []string{}
		for _, v := range values {
			selectAction := v.(map[string]interface{})["multi_users_select-action"]
			if selectAction == nil {
				continue
			}
			selectedUsers := selectAction.(map[string]interface{})["selected_users"].([]interface{})
			for _, u := range selectedUsers {
				users = append(users, u.(string))
			}
		}

		if err := svc.sendReplyDirectMessage(notifier, model.SlackDirectMsgOption{
			MentionRecordID: mentionID,
			ServiceName:     svc.Name,
			User:            msg["user"].(string),
			LinkChannel:     record.Channel,
			LinkTimestamp:   record.Timestamp,
			EventContent:    msg["text"].(string),
			Members:         users,
			ResendUserID:    resendUserID,
		}); err != nil {
			svc.l.Errorf("send reply direct message failed, err: %+v", err)
		}
	}()

	return svc.closeViewReply()
}

func (svc *SlackInteraction) closeViewReply() interface{} {
	return struct {
		ResponseAction string `json:"response_action"`
	}{
		ResponseAction: "clear",
	}
}

func (svc *SlackInteraction) homeHandler(payload map[string]interface{}) interface{} {
	action, err := svc.parseAction(payload)
	if err != nil {
		svc.l.Errorf("parse action failed, err: %+v", err)
		return svc.noneInteractionReply(payload)
	}

	switch action {
	case "clear":
		return svc.clearReply(payload)
	case "set":
		return svc.setReply(payload)
	}

	svc.l.Warn("mismatch home interactive action")
	return svc.noneInteractionReply(payload)
}

func (svc *SlackInteraction) clearReply(payload map[string]interface{}) interface{} {
	channel, err := svc.getDirectChannel(payload["user"].(map[string]interface{})["id"].(string), svc.Token)
	if err != nil || len(channel) == 0 {
		svc.l.Errorf("get channel failed, err: %+v", err)
		return svc.noneInteractionReply(payload)
	}

	notifier := util.NewSlackNotifier(svc.Token)
	path := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s", channel)
	res, _, err := notifier.Send(svc.ctx, http.MethodGet, util.Url(path), &util.GeneralMsg{})
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	json.Unmarshal(res, &data)
	if err != nil {
		svc.l.Errorf("parse data failed, err: %+v", err)
		return svc.noneInteractionReply(payload)
	}

	if data["messages"] == nil {
		svc.l.Warn("empty messages")
		return svc.noneInteractionReply(payload)
	}

	for _, msg := range data["messages"].([]interface{}) {
		ts := msg.(map[string]interface{})["ts"].(string)
		_, _, err := notifier.Send(svc.ctx, http.MethodPost, util.Url("https://slack.com/api/chat.delete"), &util.GeneralMsg{
			Channel: channel,
			TS:      ts,
		})

		if err != nil {
			return err
		}
	}

	return svc.noneInteractionReply(payload)
}

func (svc *SlackInteraction) setReply(payload map[string]interface{}) interface{} {
	svc.l.Debug("execute set")
	go func() {
		notifier := util.NewSlackNotifier(svc.Token)
		_, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostView,
			svc.settingView(payload["trigger_id"].(string)),
		)
		if err != nil {
			svc.l.Errorf("send set action view failed, err: %+v", err)
			return
		}
	}()

	return svc.noneInteractionReply(payload)
}

func (svc *SlackInteraction) settingView(triggerID string) util.SlackViewMsg {
	return util.SlackViewMsg{
		TriggerID: triggerID,
		View: map[string]interface{}{
			"type": "modal",
			"submit": util.PlainText{
				Type:  "plain_text",
				Text:  "確認",
				Emoji: true,
			},
			"close": util.PlainText{
				Type:  "plain_text",
				Text:  "取消",
				Emoji: true,
			},
			"title": util.PlainText{
				Type:  "plain_text",
				Text:  "更改機器人設定",
				Emoji: true,
			},
			"blocks": fmt.Sprintf(`[
				{
					"type": "header",
					"text": {
						"type": "plain_text",
						"text": "輪值設定",
						"emoji": true
					}
				},
				{
					"type": "divider"
				},
				{
					"type": "input",
					"element": {
						"type": "multi_users_select",
						"placeholder": {
							"type": "plain_text",
							"text": "Select users",
							"emoji": true
						},
						"action_id": "multi_users_select-action",
						"initial_users": [
							%s
						]
					},
					"label": {
						"type": "plain_text",
						"text": "輪值人員",
						"emoji": true
					}
				},
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "*開始輪值日期*"
					},
					"accessory": {
						"type": "datepicker",
						"initial_date": "1990-04-28",
						"placeholder": {
							"type": "plain_text",
							"text": "Select a date",
							"emoji": true
						},
						"action_id": "datepicker-action"
					}
				},
				{
					"type": "header",
					"text": {
						"type": "plain_text",
						"text": "訊息設定",
						"emoji": true
					}
				},
				{
					"type": "divider"
				},
				{
					"type": "input",
					"element": {
						"type": "plain_text_input",
						"action_id": "plain_text_input-action",
						"initial_value": "123123\n123213",
						"multiline": true
					},
					"label": {
						"type": "plain_text",
						"text": "%s",
						"emoji": true
					}
				}
			]`,
				`"U01QCKG7529"`,
				"機器人回覆",
			),
		},
	}
}
