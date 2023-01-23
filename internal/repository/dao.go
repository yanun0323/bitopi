package repository

import (
	"bitopi/internal/domain"
	"bitopi/internal/repository/github"
	"bitopi/internal/repository/sqlite3"
)

type Repo struct {
	github.GithubDao
	sqlite3.SqlDao
}

func NewRepo() (domain.Repository, error) {
	sqlDao, err := sqlite3.New()
	if err != nil {
		return nil, err
	}
	return Repo{
		GithubDao: github.NewGithubDao(),
		SqlDao:    sqlDao,
	}, nil
}
