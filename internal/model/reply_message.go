package model

type BotMessage struct {
	ID                 uint64 `gorm:"column:id;autoIncrement"`
	Service            string `gorm:"column:service;size:50"`
	MentionMessage     string `gorm:"column:mention_message;size:255"`
	MentionMultiMember bool   `gorm:"column:mention_multi_member;not null"`
	WeeklyMessage      string `gorm:"column:weekly_message;size:255"`
}

func (BotMessage) TableName() string {
	return "slack_bot_messages"
}
