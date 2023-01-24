package domain

import (
	"time"
)

type Repository interface {
	ListMember(serviceType string) ([]string, error)
	UpdateMember(serviceType string, member []string) error

	IsAdmin(name, serviceType string) (bool, error)
	ListAdmin(serviceType string) ([]string, error)
	SetAdmin(name, serviceType string, admin bool) error

	GetStartDate(serviceType string) (time.Time, error)
	UpdateStartDate(serviceType string, t time.Time) error
}
