package model

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

type SlackTextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackMessageButton struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Style string `json:"style"` /* default, primary, danger */
	Type  string `json:"type"`
	Value string `json:"value"`
}

/*
style: default, primary, danger
*/
func NewMessageActionButton(style, value, text string) SlackMessageButton {
	return SlackMessageButton{
		Name:  "direct_msg_action",
		Text:  text,
		Style: style,
		Type:  "button",
		Value: value,
	}
}

type SlackPermalinkRequest struct {
	Token            string `json:"token"`
	Channel          string `json:"channel"`
	MessageTimestamp string `json:"message_ts"`
}

type SlackPermalinkResponse struct {
	OK        bool   `json:"ok"`
	Channel   string `json:"channel"`
	Permalink string `json:"permalink"`
	Error     string `json:"error"`
}

type SlackActionPayload struct {
	Payload map[string]interface{} `json:"payload"`
}
