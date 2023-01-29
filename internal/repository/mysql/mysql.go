package mysql

import (
	"bitopi/internal/model"
	"fmt"
	"time"

	"github.com/pkg/errors"

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

	if err := migrate(db, &model.MentionRecord{}); err != nil {
		return err
	}

	if err := migrate(db, &model.ReplyMessage{}); err != nil {
		return err
	}

	return nil
}

func migrate(db *gorm.DB, p interface{}) error {
	if db.Migrator().HasTable(p) {
		return nil
	}

	return db.AutoMigrate(p)
}

func notFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (dao MysqlDao) ListMember(service string) ([]string, error) {
	var member []model.Member
	err := dao.db.Where("`service` = ?", service).
		Find(&member).Error
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(member))
	for _, m := range member {
		result = append(result, m.UserID)
	}
	return result, nil
}

func (dao MysqlDao) UpdateMember(service string, member []model.Member) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// FIXME: need to query first and delete object by query result
		if err := tx.Where("`service` = ?", service).
			Where("`admin` <> ?", true).
			Delete(&model.Member{}).Error; err != nil {
			return err
		}

		members := make([]model.Member, 0, len(member))
		for i, m := range member {
			members = append(members, model.Member{
				UserID:   m.UserID,
				UserName: m.UserName,
				Order:    i,
				Service:  service,
			})
		}

		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		return nil
	})
}

func (dao MysqlDao) IsAdmin(name, service string) (bool, error) {
	var count int64
	err := dao.db.Model(&model.Member{}).
		Where("`name` = ?", name).
		Where("`service` = ?", service).
		Where("`admin` = ?", true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao MysqlDao) ListAdmin(service string) ([]string, error) {
	var members []model.Member
	err := dao.db.
		Where("`service` = ?", service).
		Where("`admin` = ?", true).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(members))
	for i := range members {
		result = append(result, members[i].UserID)
	}
	return result, nil
}

func (dao MysqlDao) SetAdmin(name, service string, admin bool) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		var member model.Member

		err := tx.Where("`name` = ?", name).
			Where("`service` = ?", service).
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

func (dao MysqlDao) GetStartDate(service string) (time.Time, error) {
	elem := model.StartTime{}
	err := dao.db.Where("`service` = ?", service).
		First(&elem).Error
	if err != nil {
		return time.Time{}, err
	}

	return elem.StartTime, nil
}

func (dao MysqlDao) UpdateStartDate(service string, t time.Time) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		elem := model.StartTime{}
		err := tx.Where("`service` = ?", service).
			First(&elem).Error
		if notFound(err) {
			elem.Service = service
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

func (dao MysqlDao) FindOrCreateMentionRecord(service, channel, timestamp string) (bool, error) {
	found := false
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		record := model.MentionRecord{}
		err := tx.Where("`service` = ?", service).
			Where("`channel` = ?", channel).
			Where("`timestamp` = ?", timestamp).
			First(&record).Error
		if err == nil {
			found = true
			return nil
		}

		if !notFound(err) {
			return errors.Wrap(err, "query")
		}

		record.Service = service
		record.Channel = channel
		record.Timestamp = timestamp
		record.CreateAtu = time.Now().Unix()
		if err := tx.Create(&record).Error; err != nil {
			return errors.Wrap(err, "create")
		}
		return nil
	})
	return found, err
}

func (dao MysqlDao) GetReplyMessage(service string) (model.ReplyMessage, error) {
	msg := model.ReplyMessage{}
	if err := dao.db.Where("`service` = ?", service).First(&msg).Error; err != nil && !notFound(err) {
		return model.ReplyMessage{}, err
	}
	return msg, nil
}

func (dao MysqlDao) SetReplyMessage(service, message string, multiMember bool) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		msg := model.ReplyMessage{}
		err := tx.Where("`service` = ?", service).First(&msg).Error
		if err == nil {
			msg.Message = message
			msg.MultiMember = multiMember
			return tx.Save(&msg).Error
		}

		if !notFound(err) {
			return err
		}

		msg = model.ReplyMessage{
			Service:     service,
			Message:     message,
			MultiMember: multiMember,
		}

		return tx.Create(&msg).Error
	})
}
