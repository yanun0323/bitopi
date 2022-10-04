package sqlite3

import (
	"bitopi/intrernal/model"
	"time"

	"github.com/pkg/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlDao struct {
	db *gorm.DB
}

func New() (SqlDao, error) {
	db, err := gorm.Open(sqlite.Open("bitopi.db"), &gorm.Config{})
	if err != nil {
		return SqlDao{}, err
	}

	if err := autoMigrate(db); err != nil {
		return SqlDao{}, err
	}

	return SqlDao{
		db: db,
	}, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.Maid{}, &model.Admin{}, &model.Setting{})
}

func (dao SqlDao) ListMaid() ([]string, error) {
	var maids []model.Maid
	err := dao.db.Find(&maids).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(maids))
	for i := range maids {
		result = append(result, maids[i].Name)
	}
	return result, nil
}

func (dao SqlDao) UpdateMaidList(list []string) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM maids").Error
		if err != nil {
			return err
		}

		for i := range list {
			elem := model.Maid{
				Name:  list[i],
				Order: i,
			}

			err := tx.Create(&elem).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao SqlDao) IsAdmin(admin string) (bool, error) {
	var count int64
	err := dao.db.Model(&model.Admin{}).
		Where("name=?", admin).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (dao SqlDao) ListAdmin() ([]string, error) {
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

func (dao SqlDao) ReverseAdmin(admin string) error {
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

func (dao SqlDao) GetStartDate() (time.Time, error) {
	elem := model.Setting{}
	err := dao.db.First(&elem).Error
	if err != nil {
		return time.Now(), err
	}

	return elem.StartTime, nil
}

func (dao SqlDao) UpdateStartDate(t time.Time) error {
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
