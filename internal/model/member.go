package model

type Member struct {
	ID       uint64 `gorm:"column:id;autoIncrement;primaryKey" json:"-"`
	UserID   string `gorm:"column:user_id;size:50" json:"user_id"`
	UserName string `gorm:"column:user_name;size:50" json:"user_name"`
	Order    int    `gorm:"column:order" json:"order"`
	Service  string `gorm:"column:service;size:50" json:"-"`
}

func (Member) TableName() string {
	return "slack_bot_members"
}

func (m Member) UserTag() string {
	return "<@" + m.UserID + ">"
}
