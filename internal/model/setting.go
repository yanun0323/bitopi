package model

import "time"

type Setting struct {
	ID        uint64    `gorm:"column:id,autoIncrement"`
	StartTime time.Time `gorm:"start_time"`
}

func (s Setting) TableName() string {
	return "settings"
}
