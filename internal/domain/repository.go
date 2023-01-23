package domain

import (
	"time"
)

type Repository interface {
	ListMember(string) ([]string, error)
	UpdateMember(string, []string) error

	// TODO: Refactor for different team
	IsAdmin(admin string) (bool, error)
	ListAdmin() ([]string, error)
	ReverseAdmin(admin string) error

	// TODO: Refactor for different team
	GetStartDate() (time.Time, error)
	UpdateStartDate(t time.Time) error
}
