package service

import (
	"bitopi/internal/util"
	"fmt"
	"net/http"
	"strings"
)

func (svc *SlackBot) publishHomeView(notifier util.SlackNotifier, userMention string) error {
	res, _, err := notifier.Send(svc.ctx, http.MethodPost, util.PostHome, svc.homeView(userMention))
	if err != nil {
		return err
	}
	svc.l.Debugf("home view response:\n%s", string(res))
	return nil
}

// TODO: Bot Home: updating App home information
// after every api called and interactor and cronJob
func (svc *SlackBot) homeView(userMention string) util.SlackHomeViewMsg {
	times, err := svc.repo.CountMentionRecord(svc.Name)
	if err != nil {
		svc.l.Errorf("count mention record error, %+v", err)
		return util.SlackHomeViewMsg{}
	}

	dutyMember, _, err := svc.getDutyMember(false)
	if err != nil {
		svc.l.Errorf("get duty member error, %+v", err)
		return util.SlackHomeViewMsg{}
	}

	members, err := svc.listMember(true)
	if err != nil {
		svc.l.Errorf("list members error, %+v", err)
		return util.SlackHomeViewMsg{}
	}

	history := `*更新歷史*
- 2023.1 新增私訊通知功能/新增首頁按鈕
`

	return util.SlackHomeViewMsg{
		UserID: userMention[2 : len(userMention)-1],
		View: map[string]interface{}{
			"type": "home",
			"blocks": fmt.Sprintf(`[
					{
						"type": "section",
						"text": {
							"type": "mrkdwn",
							"text": "*本週輪值人員* \n<@%s> \n\n*輪值人員順序* \n%s"
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
						"elements": [
							{
								"type": "button",
								"text": {
									"type": "plain_text",
									"text": "清除所有私訊",
									"emoji": true
								},
								"style": "danger",
								"value": "clear",
								"action_id": "clear",
								"confirm": {
									"style": "danger",
									"title": {
										"type": "plain_text",
										"text": "清除所有私訊"
									},
									"text": {
										"type": "mrkdwn",
										"text": "是否清除此機器人傳送給您的所有訊息？"
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
				dutyMember,
				strings.Join(members, " "),
				times,
				history,
			),
		},
	}
}

// TODO: Add to home view and validate is administrator
/*
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
	},
*/
