package model

type SlackTextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackDirectMsgOption struct {
	IsUser          bool
	MentionRecordID string
	ServiceName     string
	User            string
	EventContent    string
	Members         []string
	ResendUserID    string
	LinkChannel     string
	LinkTimestamp   string
}
