package domain

import (
	"bitopi/internal/model"
	"context"
	"time"
)

type Repository interface {
	Tx(ctx context.Context, fn func(context.Context) error) error

	GetMember(ctx context.Context, service string, userID string) (model.Member, error)
	UpdateMember(ctx context.Context, member model.Member) error

	ListMembers(ctx context.Context, service string) ([]model.Member, error)
	ResetMembers(txCtx context.Context, service string, member []model.Member) error

	ListAllMembers(ctx context.Context) ([]model.Member, error)

	IsAdmin(ctx context.Context, service, userID string) (bool, error)
	ListAdmin(ctx context.Context, service string) ([]model.Admin, error)
	AddAdmin(ctx context.Context, admin model.Admin) error
	DeleteAdmin(ctx context.Context, service, userID string) error

	GetStartDate(ctx context.Context, service string) (time.Time, error)
	UpdateStartDate(txCtx context.Context, service string, t time.Time) error

	GetDutyDuration(ctx context.Context, service string) (time.Duration, error)
	GetDutyMemberCountPerTime(ctx context.Context, service string) (int, error)

	CountMentionRecord(ctx context.Context, service string) (int64, error)
	GetMentionRecord(ctx context.Context, id uint64) (model.MentionRecord, error)
	FindOrCreateMentionRecord(txCtx context.Context, service, channel, timestamp string) (id uint64, found bool, err error)

	GetReplyMessage(ctx context.Context, service string) (model.BotMessage, error)
	SetReplyMessage(txCtx context.Context, msg model.BotMessage) error

	GetSubscriber(ctx context.Context) ([]model.Subscriber, error)
	SetSubscriber(ctx context.Context, sub model.Subscriber) error
	DeleteSubscriber(ctx context.Context, sub model.Subscriber) error
}
