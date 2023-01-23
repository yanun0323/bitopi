package model

const MemberTableName = "slack_bot_members"

type Member struct {
	ID    uint64 `gorm:"column:id,autoIncrement"`
	Name  string `gorm:"column:name;size:50"`
	Order int    `gorm:"column:order"`
	Type  string `gorm:"column:type;size:50"`
	Admin bool   `gorm:"column:admin;not null;default:true"`
}

func (Member) TableName() string {
	return MemberTableName
}
