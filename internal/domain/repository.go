package domain

import (
	"bitopi/internal/model"
	"time"
)

type Repository interface {
	ListMember(service string) ([]model.Member, error)
	UpdateMember(service string, member []model.Member) error

	// IsAdmin(name, service string) (bool, error)
	// ListAdmin(service string) ([]string, error)
	// SetAdmin(name, service string, admin bool) error

	GetStartDate(service string) (time.Time, error)
	UpdateStartDate(service string, t time.Time) error

	CountMentionRecord(service string) (int64, error)
	GetMentionRecord(id uint64) (model.MentionRecord, error)
	FindOrCreateMentionRecord(service, channel, timestamp string) (id uint64, found bool, err error)

	GetReplyMessage(service string) (model.BotMessage, error)
	SetReplyMessage(msg model.BotMessage) error
}
