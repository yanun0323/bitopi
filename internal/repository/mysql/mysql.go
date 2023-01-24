package mysql

import (
	"bitopi/internal/model"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDao struct {
	db *gorm.DB
}

func New() (MysqlDao, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.username"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.database"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return MysqlDao{}, err
	}

	if err := initMigration(db); err != nil {
		return MysqlDao{}, err
	}

	return MysqlDao{
		db: db,
	}, nil
}

func initMigration(db *gorm.DB) error {
	if err := migrate(db, &model.Member{}); err != nil {
		return err
	}

	if err := migrate(db, &model.StartTime{}); err != nil {
		return err
	}

	return nil
}

func migrate(db *gorm.DB, p interface{}) error {
	if db.Migrator().HasTable(p) {
		return nil
	}

	if err := db.AutoMigrate(p); err != nil {
		return err
	}

	return nil
}

func notFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (dao MysqlDao) ListMember(serviceType string) ([]string, error) {
	var member []model.Member
	err := dao.db.Where("`service` = ?", serviceType).
		Find(&member).Error
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(member))
	for _, m := range member {
		result = append(result, m.Name)
	}
	return result, nil
}

func (dao MysqlDao) UpdateMember(serviceType string, member []string) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// FIXME: need to query first and delete object by query result
		if err := tx.Where("`service` = ?", serviceType).
			Where("`admin` <> ?", true).
			Delete(&model.Member{}).Error; err != nil {
			return err
		}

		members := make([]model.Member, 0, len(member))
		for i, m := range member {
			members = append(members, model.Member{
				Name:    m,
				Order:   i,
				Service: serviceType,
			})
		}

		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		return nil
	})
}

func (dao MysqlDao) IsAdmin(name, serviceType string) (bool, error) {
	var count int64
	err := dao.db.Model(&model.Member{}).
		Where("`name` = ?", name).
		Where("`service` = ?", serviceType).
		Where("`admin` = ?", true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao MysqlDao) ListAdmin(serviceType string) ([]string, error) {
	var members []model.Member
	err := dao.db.
		Where("`service` = ?", serviceType).
		Where("`admin` = ?", true).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(members))
	for i := range members {
		result = append(result, members[i].Name)
	}
	return result, nil
}

func (dao MysqlDao) SetAdmin(name, serviceType string, admin bool) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		var member model.Member

		err := tx.Where("`name` = ?", name).
			Where("`service` = ?", serviceType).
			Where("`order` = ?", -1).
			First(&member).Error
		if err != nil && !notFound(err) {
			return err
		}

		if !admin {
			return nil
		}

		member.Admin = admin
		err = tx.Save(&member).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (dao MysqlDao) GetStartDate(serviceType string) (time.Time, error) {
	elem := model.StartTime{}
	err := dao.db.Where("`service` = ?", serviceType).
		First(&elem).Error
	if err != nil {
		return time.Time{}, err
	}

	return elem.StartTime, nil
}

func (dao MysqlDao) UpdateStartDate(serviceType string, t time.Time) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		elem := model.StartTime{}
		err := tx.Where("`service` = ?", serviceType).
			First(&elem).Error
		if notFound(err) {
			elem.Service = serviceType
			elem.StartTime = t
			return tx.Create(&elem).Error
		}

		if err != nil {
			return err
		}

		elem.StartTime = t
		err = tx.Save(&elem).Error
		if err != nil {
			return err
		}

		return nil
	})
}
