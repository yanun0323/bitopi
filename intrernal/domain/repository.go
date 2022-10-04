package domain

import "time"

type IRepository interface {
	ListMaid() ([]string, error)
	UpdateMaidList(list []string) error

	IsAdmin(admin string) (bool, error)
	ListAdmin() ([]string, error)
	ReverseAdmin(admin string) error

	GetStartDate() (time.Time, error)
	UpdateStartDate(t time.Time) error
}
