package model

type BotSetting struct {
	ID    uint64 `gorm:"column:id;autoIncrement"`
	Key   string `gorm:"column:key"`
	Value string `gorm:"column:value"`
}

func (BotSetting) TableName() string {
	return "slack_bot_settings"
}
