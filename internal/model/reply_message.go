package model

type ReplyMessage struct {
	ID          uint64 `gorm:"column:id;autoIncrement"`
	Service     string `gorm:"column:service;size:50"`
	Message     string `gorm:"column:name;size:50"`
	MultiMember bool   `gorm:"column:multi_member;not null"`
}

func (ReplyMessage) TableName() string {
	return "slack_reply_messages"
}
