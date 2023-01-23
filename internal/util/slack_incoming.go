package util

type SlackVerificationResponse struct {
	Challenge string `json:"challenge"`
}

type SlackTypeCheck struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}

type SlackVerification struct {
	Challenge string `json:"challenge"`
}

type SlackEventAPI struct {
	Token       string   `json:"token"`
	TeamID      string   `json:"team_id"`
	ApiAppID    string   `json:"api_app_id"`
	Event       Event    `json:"event"`
	Type        string   `json:"type"`
	EventId     string   `json:"event_id"`
	EventTime   uint64   `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
}

type Event struct {
	Type           string `json:"type"`
	User           string `json:"user"`
	Text           string `json:"text"`
	TimeStamp      string `json:"ts"`
	Channel        string `json:"channel"`
	EventTimeStamp string `json:"event_ts"`
}

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
	Type            string       `json:"type"`
	Token           string       `json:"token"`
	ActionTimeStamp string       `json:"action_ts"`
	Team            ShortcutTeam `json:"team"`
	User            ShortcutUser `json:"user"`
	CallbackId      string       `json:"callback_id"`
	TriggerId       string       `json:"trigger_id"`
}

type ShortcutTeam struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type ShortcutUser struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	TeamID   string `json:"team_id"`
}
