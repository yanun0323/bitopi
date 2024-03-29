package util

import (
	"encoding/json"
)

type Messenger interface {
	Marshal() ([]byte, error)
}

type SlackMapMsg map[string]interface{}

func (msg SlackMapMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type GeneralMsg struct {
	Text    string `json:"text,omitempty"`
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
}

func (msg GeneralMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type SlackMsg struct {
	Text        string              `json:"text"`
	Channel     string              `json:"channel"`
	UserName    string              `json:"username,omitempty"`
	Attachments []map[string]string `json:"attachments,omitempty"`
}

func (msg *SlackMsg) AddAttachments(vs ...string) *SlackMsg {
	if cap(msg.Attachments) == 0 {
		msg.Attachments = make([]map[string]string, 0, 1)
	}
	data := make(map[string]string)
	for i := 0; i < len(vs); i += 2 {
		data[vs[i]] = vs[i+1]
	}
	msg.Attachments = append(msg.Attachments, data)
	return msg
}

func (msg SlackMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type SlackReplyMsg struct {
	Text        string                   `json:"text"`
	Channel     string                   `json:"channel"`
	UserName    string                   `json:"username,omitempty"`
	TimeStamp   string                   `json:"thread_ts"`
	UnfurlLinks bool                     `json:"unfurl_links"`
	UnfurlMedia bool                     `json:"unfurl_media"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
}

func (msg SlackReplyMsg) AddAttachments(vs ...interface{}) SlackReplyMsg {
	if cap(msg.Attachments) == 0 {
		msg.Attachments = make([]map[string]interface{}, 0, 1)
	}
	data := make(map[string]interface{})
	for i := 0; i < len(vs); i += 2 {
		data[vs[i].(string)] = vs[i+1]
	}
	msg.Attachments = append(msg.Attachments, data)
	return msg
}

func (msg SlackReplyMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type SlackPermalinkRequest struct {
	Channel          string `json:"channel"`
	MessageTimestamp string `json:"message_ts"`
}

func (msg SlackPermalinkRequest) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type SlackHomeViewMsg struct {
	UserID string                 `json:"user_id"`
	View   map[string]interface{} `json:"view"`
}

func (msg SlackHomeViewMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type SlackViewMsg struct {
	TriggerID string                 `json:"trigger_id"`
	View      map[string]interface{} `json:"view"`
}

type PlainText struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}

func (msg SlackViewMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

func GetInteractor(triggerID string) SlackViewMsg {
	return SlackViewMsg{
		TriggerID: triggerID,
		View: map[string]interface{}{
			"type": "modal",
			"submit": PlainText{
				Type:  "plain_text",
				Text:  "Submit",
				Emoji: true,
			},
			"close": PlainText{
				Type:  "plain_text",
				Text:  "Cancel",
				Emoji: true,
			},
			"title": PlainText{
				Type:  "plain_text",
				Text:  "Ticket app",
				Emoji: true,
			},
			"blocks": `[
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "女僕名單",
					"emoji": true
				}
			},
			{
				"type": "divider"
			},
			{
				"type": "section",
				"text": {
					"type": "plain_text",
					"text": "<@U032TJB1PE1>"
				},
				"accessory": {
					"type": "overflow",
					"options": [
						{
							"text": {
								"type": "plain_text",
								"text": ":pencil: Up",
								"emoji": true
							},
							"value": "Up"
						},
						{
							"text": {
								"type": "plain_text",
								"text": ":pencil: Down",
								"emoji": true
							},
							"value": "Down"
						},
						{
							"text": {
								"type": "plain_text",
								"text": ":x: Delete",
								"emoji": true
							},
							"value": "delete"
						}
					]
				}
			},
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "*<fakelink.com|MOB-2011 Deep-link from web search results to product page>*"
				},
				"accessory": {
					"type": "overflow",
					"options": [
						{
							"text": {
								"type": "plain_text",
								"text": ":pencil: Up",
								"emoji": true
							},
							"value": "Up"
						},
						{
							"text": {
								"type": "plain_text",
								"text": ":pencil: Down",
								"emoji": true
							},
							"value": "Down"
						},
						{
							"text": {
								"type": "plain_text",
								"text": ":x: Delete",
								"emoji": true
							},
							"value": "delete"
						}
					]
				}
			},
			{
				"dispatch_action": true,
				"type": "input",
				"element": {
					"type": "plain_text_input",
					"action_id": "plain_text_input-action"
				},
				"label": {
					"type": "plain_text",
					"text": "新增女僕",
					"emoji": true
				}
			},
			{
				"type": "input",
				"element": {
					"type": "datepicker",
					"initial_date": "1995-03-23",
					"placeholder": {
						"type": "plain_text",
						"text": "Select a date",
						"emoji": true
					},
					"action_id": "datepicker-action"
				},
				"label": {
					"type": "plain_text",
					"text": "開始輪流日期",
					"emoji": true
				}
			}
		]`,
		},
	}
}
