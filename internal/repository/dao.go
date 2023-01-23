package repository

import (
	"bitopi/internal/domain"
	"bitopi/internal/repository/mysql"
)

type Repo struct {
	mysql.MysqlDao
}

func NewRepo() (domain.Repository, error) {
	mysqlDao, err := mysql.New()
	if err != nil {
		return nil, err
	}
	return Repo{
		MysqlDao: mysqlDao,
	}, nil
}
