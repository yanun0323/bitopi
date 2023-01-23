package model

type Admin struct {
	ID   uint64 `gorm:"column:id,autoIncrement"`
	Name string `gorm:"column:name"`
}

func (a Admin) TableName() string {
	return "admins"
}
