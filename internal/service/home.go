package service

import (
	"bitopi/internal/util"
	"fmt"
	"net/http"
	"strings"
)

func (svc *SlackBot) publishHomeView(notifier util.SlackNotifier) error {
	subscribers, err := svc.repo.GetSubscriber()
	if err != nil {
		return err
	}

	members, err := svc.repo.ListAllMembers()
	if err != nil {
		return err
	}

	subscriberIDs := map[string]bool{}
	for _, subscriber := range subscribers {
		subscriberIDs[subscriber.UserID] = true
	}
	for _, member := range members {
		subscriberIDs[member.UserID] = true
	}

	view, err := svc.getHomeView(false)
	if err != nil {
		return err
	}

	for subscriberID := range subscriberIDs {
		if _, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostHome, svc.createHomeViewRequest(view, subscriberID)); err != nil {
			return err
		}
	}

	return nil
}

func (svc *SlackBot) createHomeViewRequest(view map[string]interface{}, userID string) *util.SlackHomeViewMsg {
	return &util.SlackHomeViewMsg{
		UserID: userID,
		View:   view,
	}
}

func (svc *SlackBot) getHomeView(isAdmin bool) (map[string]interface{}, error) {
	mentionTimes, err := svc.repo.CountMentionRecord(svc.Name)
	if err != nil {
		svc.l.Errorf("count mention record error, %+v", err)
		return nil, err
	}

	dutyMember, leftMembers, err := svc.getDutyMember(true)
	if err != nil {
		svc.l.Errorf("get duty member error, %+v", err)
		return nil, err
	}

	members, err := svc.listMember(true)
	if err != nil {
		svc.l.Errorf("list members error, %+v", err)
		return nil, err
	}

	rMsg, err := svc.repo.GetReplyMessage(svc.Name)
	if err != nil {
		svc.l.Errorf("get reply message error, %+v")
		return nil, err
	}

	replyText := ""
	if rMsg.MentionMultiMember {
		replyText = fmt.Sprintf(rMsg.HomeMentionMessage, dutyMember, strings.Join(leftMembers, " "))
	} else {
		replyText = fmt.Sprintf(rMsg.HomeMentionMessage, dutyMember)
	}

	history := `*更新歷史*
- 2023.1 新增私訊通知功能/新增首頁按鈕
`

	adminSetButton := ""
	if isAdmin {
		adminSetButton = `
{
	"type": "button",
	"text": {
		"type": "plain_text",
		"text": "更改設定",
		"emoji": true
	},
	"style": "primary",
	"value": "set",
	"action_id": "set"
},`
	}

	return map[string]interface{}{
		"type": "home",
		"blocks": fmt.Sprintf(`[
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "%s \n\n*輪值人員順序* \n%s"
					}
				},
				{
					"type": "context",
					"elements": [
						{
							"type": "mrkdwn",
							"text": "此機器人已被提及 %d 次"
						}
					]
				},
				{
					"type": "actions",
					"elements": [%s
						{
							"type": "button",
							"text": {
								"type": "plain_text",
								"text": "刪除所有通知",
								"emoji": true
							},
							"style": "danger",
							"value": "clear",
							"action_id": "clear",
							"confirm": {
								"style": "danger",
								"title": {
									"type": "plain_text",
									"text": "刪除所有通知"
								},
								"text": {
									"type": "mrkdwn",
									"text": "是否刪除此機器人傳送給您的所有通知訊息？"
								},
								"confirm": {
									"type": "plain_text",
									"text": "清除"
								},
								"deny": {
									"type": "plain_text",
									"text": "取消"
								}
							}
						}
					]
				},
				{
					"type": "context",
					"elements": [
						{
							"type": "mrkdwn",
							"text": "%s"
						}
					]
				}
			]`,
			replyText,
			strings.Join(members, " "),
			mentionTimes,
			adminSetButton,
			history,
		),
	}, nil
}
