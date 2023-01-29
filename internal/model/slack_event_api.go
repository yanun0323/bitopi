package model

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
