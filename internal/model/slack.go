package model

type SlackTextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackDirectMsgOption struct {
	IsUser          bool
	MentionRecordID string
	ServiceName     string
	Channel         string
	User            string
	EventTimestamp  string
	EventContent    string
	Members         []string
}
