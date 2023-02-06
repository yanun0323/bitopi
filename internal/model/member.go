package model

type Member struct {
	ID       uint64 `gorm:"column:id;autoIncrement"`
	UserID   string `gorm:"column:user_id;size:50"`
	UserName string `gorm:"column:user_name;size:50"`
	Order    int    `gorm:"column:order"`
	Service  string `gorm:"column:service;size:50"`
}

func (Member) TableName() string {
	return "slack_bot_members"
}

func (m Member) UserTag() string {
	return "<@" + m.UserID + ">"
}
