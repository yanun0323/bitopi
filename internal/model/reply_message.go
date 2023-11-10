package model

type BotMessage struct {
	ID                 uint64 `gorm:"column:id;autoIncrement;primaryKey" json:"-"`
	Service            string `gorm:"column:service;size:50" json:"-"`
	MentionMessage     string `gorm:"column:mention_message;size:255" json:"mention_message"`
	MentionMultiMember bool   `gorm:"column:mention_multi_member;not null" json:"mention_multi_member"`
	HomeMentionMessage string `gorm:"column:home_mention_message;size:255" json:"home_mention_message"`
	DoneReplyMessage   string `gorm:"column:done_reply_message;size:255" json:"done_reply_message"`
}

func (BotMessage) TableName() string {
	return "slack_bot_messages"
}
