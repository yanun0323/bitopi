package domain

import (
	"bitopi/internal/model"
	"time"
)

type Repository interface {
	ListMember(service string) ([]string, error)
	UpdateMember(service string, member []model.Member) error

	IsAdmin(name, service string) (bool, error)
	ListAdmin(service string) ([]string, error)
	SetAdmin(name, service string, admin bool) error

	GetStartDate(service string) (time.Time, error)
	UpdateStartDate(service string, t time.Time) error

	FindOrCreateMentionRecord(service, channel, timestamp string) (found bool, err error)

	GetReplyMessage(service string) (model.ReplyMessage, error)
	SetReplyMessage(service, message string, multiMember bool) error
}
