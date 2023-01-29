package model

/*
Json example:

	`
	{
		"type": "shortcut",
		"token": "XXXXXXXXXXXXX",
		"action_ts": "1581106241.371594",
		"team": {
		"id": "TXXXXXXXX",
		"domain": "shortcuts-test"
		},
		"user": {
		"id": "UXXXXXXXXX",
		"username": "aman",
		"team_id": "TXXXXXXXX"
		},
		"callback_id": "shortcut_create_task",
		"trigger_id": "944799105734.773906753841.38b5894552bdd4a780554ee59d1f3638"
	}
	`
*/
type SlackShortcutPayload struct {
	Type            string            `json:"type"`
	Token           string            `json:"token"`
	ActionTimeStamp string            `json:"action_ts"`
	Team            SlackShortcutTeam `json:"team"`
	User            SlackShortcutUser `json:"user"`
	CallbackId      string            `json:"callback_id"`
	TriggerId       string            `json:"trigger_id"`
}

type SlackShortcutTeam struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type SlackShortcutUser struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	TeamID   string `json:"team_id"`
}
