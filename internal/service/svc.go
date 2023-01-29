package service

import (
	"bitopi/internal/domain"
	"bitopi/internal/repository"
	"context"

	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/logs"
)

type Service struct {
	repo     domain.Repository
	l        *logs.Logger
	ctx      context.Context
	logLevel uint8
}

func New() (Service, error) {
	repo, err := repository.NewRepo()
	if err != nil {
		return Service{}, err
	}
	ctx := context.Background()
	logLevel := uint8(viper.GetUint16("log.level"))
	return Service{
		repo:     repo,
		l:        logs.New("bito_pi", logLevel),
		ctx:      ctx,
		logLevel: logLevel,
	}, nil
}
