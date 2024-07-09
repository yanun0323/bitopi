package model

type SlackDirectChannel struct {
	OK      bool         `json:"ok"`
	Channel SlackChannel `json:"channel"`
}

type SlackChannel struct {
	ID string `json:"id"`
}
