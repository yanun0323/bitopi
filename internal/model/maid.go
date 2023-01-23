package model

type Maid struct {
	ID    uint64 `gorm:"column:id,autoIncrement"`
	Name  string `gorm:"column:name"`
	Order int    `gorm:"column:order"`
}

func (m Maid) TableName() string {
	return "maids"
}
