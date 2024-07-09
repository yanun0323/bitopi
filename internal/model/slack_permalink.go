package model

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
