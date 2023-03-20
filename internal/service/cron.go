package service

import "bitopi/internal/util"

type WeeklyNotifier struct {
	SlackBot
	WeeklyNotifierOpt
}
type WeeklyNotifierOpt struct{}

func NewWeeklyJob(bot SlackBot, opt WeeklyNotifierOpt) *WeeklyNotifier {
	return &WeeklyNotifier{
		SlackBot:          bot,
		WeeklyNotifierOpt: opt,
	}
}

func (svc *WeeklyNotifier) Run() {
	notifier := util.NewSlackNotifier(svc.SlackBot.Token)
	err := svc.SlackBot.publishHomeView(notifier)
	if err != nil {
		svc.l.Errorf("publish home view tab failed, err: %+v", err)
		return
	}
}
