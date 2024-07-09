package model

type Admin struct {
	ID       uint64 `gorm:"column:id;autoIncrement"`
	UserID   string `gorm:"column:user_id;size:50"`
	UserName string `gorm:"column:user_name;size:50"`
	Service  string `gorm:"column:service;size:50"`
}

func (Admin) TableName() string {
	return "slack_bot_admins"
}

func (a Admin) IsEmpty() bool {
	return len(a.Service) == 0 || len(a.UserID) == 0
}
