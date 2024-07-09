package model

import "time"

type StartTime struct {
	ID        uint64    `gorm:"column:id;autoIncrement"`
	Service   string    `gorm:"column:service;size:50"`
	StartTime time.Time `gorm:"start_time"`
}

func (s StartTime) TableName() string {
	return "slack_bot_start_time"
}
