package model

type Member struct {
	ID    uint64 `gorm:"column:id,autoIncrement"`
	Name  string `gorm:"column:name"`
	Order int    `gorm:"column:order"`
}
