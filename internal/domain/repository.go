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

	IsAdmin(service, userID string) (bool, error)
	ListAdmin(service string) ([]model.Admin, error)
	AddAdmin(admin model.Admin) error
	DeleteAdmin(service, userID string) error

	GetStartDate(service string) (time.Time, error)
	UpdateStartDate(service string, t time.Time) error

	GetDutyDuration(service string) (time.Duration, error)
	GetDutyMemberCountPerTime(service string) (int, error)

	CountMentionRecord(service string) (int64, error)
	GetMentionRecord(id uint64) (model.MentionRecord, error)
	FindOrCreateMentionRecord(service, channel, timestamp string) (id uint64, found bool, err error)

	GetReplyMessage(service string) (model.BotMessage, error)
	SetReplyMessage(msg model.BotMessage) error

	GetSubscriber() ([]model.Subscriber, error)
	SetSubscriber(sub model.Subscriber) error
	DeleteSubscriber(sub model.Subscriber) error
}
