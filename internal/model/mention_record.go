package model

type MentionRecord struct {
	ID        uint64 `gorm:"column:id;autoIncrement"`
	Service   string `gorm:"column:service;size:50;index;not null"`
	Channel   string `gorm:"column:channel;size:50;index;not null"`
	Timestamp string `gorm:"column:timestamp;size:50;index;not null"`
	CreateAtu int64  `gorm:"column:create_atu;not null"`
}

func (s MentionRecord) TableName() string {
	return "slack_mention_records"
}
