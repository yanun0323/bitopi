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

	if err := migrate(db, &model.Admin{}); err != nil {
		return err
	}

	if err := migrate(db, &model.Setting{}); err != nil {
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

func (dao MysqlDao) ListMember(memberType string) ([]string, error) {
	var member []model.Member
	if err := dao.db.Where("type = ?", memberType).Find(&member).Error; err != nil {
		return nil, err
	}

	result := make([]string, 0, len(member))
	for _, m := range member {
		result = append(result, m.Name)
	}
	return result, nil
}

func (dao MysqlDao) UpdateMember(memberType string, member []string) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// FIXME: need to query first and delete object by query result
		if err := tx.Where("type = ? AND order <> ?", memberType, -1).Delete(&model.Member{}).Error; err != nil {
			return err
		}

		members := make([]model.Member, 0, len(member))
		for i, m := range member {
			members = append(members, model.Member{
				Name:  m,
				Order: i,
				Type:  memberType,
			})
		}

		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		return nil
	})
}

func (dao MysqlDao) IsAdmin(admin string) (bool, error) {
	var count int64
	err := dao.db.Model(&model.Admin{}).
		Where("name=?", admin).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao MysqlDao) ListAdmin() ([]string, error) {
	var admins []model.Admin
	err := dao.db.Find(&admins).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(admins))
	for i := range admins {
		result = append(result, admins[i].Name)
	}
	return result, nil
}

func (dao MysqlDao) ReverseAdmin(admin string) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		elem := model.Admin{
			Name: admin,
		}

		err := tx.Model(&elem).
			Where("name=?", elem.Name).
			Count(&count).Error
		if err != nil {
			return err
		}

		if count > 0 {
			err := tx.Where("name = ?", elem.Name).
				Delete(&model.Admin{}).Error
			if err != nil {
				return err
			}
			return nil
		}

		err = tx.Create(&elem).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (dao MysqlDao) GetStartDate() (time.Time, error) {
	elem := model.Setting{}
	err := dao.db.First(&elem).Error
	if err != nil {
		return time.Now(), err
	}

	return elem.StartTime, nil
}

func (dao MysqlDao) UpdateStartDate(t time.Time) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		elem := model.Setting{}
		err := tx.First(&elem).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := tx.Create(&elem).Error
			if err != nil {
				return err
			}
			return nil
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
