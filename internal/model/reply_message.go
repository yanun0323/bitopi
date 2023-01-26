package model

type SlackReplyMessage struct {
	ID          uint64 `gorm:"column:id;autoIncrement"`
	Service     string `gorm:"column:service;size:50"`
	Message     string `gorm:"column:name;size:50"`
	MultiMember bool   `gorm:"column:multi_member;not null"`
}

func (SlackReplyMessage) TableName() string {
	return "slack_reply_messages"
}
