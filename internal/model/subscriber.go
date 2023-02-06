package model

type Subscriber struct {
	UserID   string `gorm:"column:user_id;size:50;primaryKey"`
	UserName string `gorm:"column:user_name;size:50"`
	Home     bool   `gorm:"column:home;size:50;not null;default:false"`
}

func (Subscriber) TableName() string {
	return "slack_bot_subscribers"
}
