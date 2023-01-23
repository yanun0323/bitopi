package model

type DirectChannel struct {
	OK      bool    `json:"ok"`
	Channel Channel `json:"channel"`
}

type Channel struct {
	ID string `json:"id"`
}
