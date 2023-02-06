package domain

import (
	"bitopi/internal/model"
	"time"
)

type Repository interface {
	GetMember(service string, userID string) (model.Member, error)
	UpdateMember(member model.Member) error

	ListMembers(service string) ([]model.Member, error)
	ResetMembers(service string, member []model.Member) error

	ListAllMembers() ([]model.Member, error)

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

	GetSubscriber() ([]model.Subscriber, error)
	SetSubscriber(sub model.Subscriber) error
	DeleteSubscriber(sub model.Subscriber) error
}
