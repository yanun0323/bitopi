package mysql

import (
	"bitopi/internal/model"
	"fmt"
	"sort"
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

	tables := []interface{}{
		&model.Member{},
		&model.StartTime{},
		&model.MentionRecord{},
		&model.BotMessage{},
		&model.Admin{},
		&model.Subscriber{},
	}

	for _, table := range tables {
		if err := migrate(db, table); err != nil {
			return err
		}
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

func (dao MysqlDao) GetMember(service string, userID string) (model.Member, error) {
	var member model.Member
	err := dao.db.
		Where("`service` = ?", service).
		Where("`user_id` = ?", userID).
		First(&member).Error
	if err != nil {
		return model.Member{}, err
	}
	return member, nil
}

func (dao MysqlDao) UpdateMember(member model.Member) error {
	var id uint64
	err := dao.db.
		Model(&model.Member{}).
		Select("`id`").
		Where("`service` = ?", member.Service).
		Where("`user_id` = ?", member.UserID).
		First(&id).Error
	if err != nil {
		return err
	}
	member.ID = id
	return dao.db.Save(&member).Error
}

func (dao MysqlDao) ListMembers(service string) ([]model.Member, error) {
	var members []model.Member
	err := dao.db.Where("`service` = ?", service).
		Order("`order`").
		Find(&members).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Order < members[j].Order
	})

	return members, nil
}

func (dao MysqlDao) ResetMembers(service string, member []model.Member) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`service` = ?", service).
			Delete(&model.Member{}).Error; err != nil && !notFound(err) {
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

func (dao MysqlDao) ListAllMembers() ([]model.Member, error) {
	var members []model.Member
	err := dao.db.Find(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (dao MysqlDao) IsAdmin(service, userID string) (bool, error) {
	err := dao.db.
		Model(&model.Admin{}).
		Where("`user_id` = ?", userID).
		Where("`service` = ?", service).Error
	if err == nil {
		return true, nil
	}

	if notFound(err) {
		return false, nil
	}

	return false, err
}

func (dao MysqlDao) ListAdmin(service string) ([]model.Admin, error) {
	var admins []model.Admin
	if err := dao.db.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (dao MysqlDao) AddAdmin(admin model.Admin) error {
	if admin.IsEmpty() {
		return errors.New(fmt.Sprintf("empty admin, %+v", admin))
	}
	return dao.db.Save(&admin).Error
}

func (dao MysqlDao) DeleteAdmin(service, userID string) error {
	err := dao.db.
		Where("`service` = ?", service).
		Where("`user_id` = ?", userID).
		Delete(&model.Admin{}).Error
	if err != nil && !notFound(err) {
		return err
	}
	return nil
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

func (dao MysqlDao) CountMentionRecord(service string) (int64, error) {
	var count int64
	if err := dao.db.Model(&model.MentionRecord{}).Where("`service` = ?", service).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (dao MysqlDao) GetMentionRecord(id uint64) (model.MentionRecord, error) {
	record := model.MentionRecord{}
	if err := dao.db.Where("`id`= ?", id).First(&record).Error; err != nil {
		return model.MentionRecord{}, err
	}
	return record, nil
}

func (dao MysqlDao) FindOrCreateMentionRecord(service, channel, timestamp string) (uint64, bool, error) {
	found := false
	var id uint64
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		record := model.MentionRecord{}
		err := tx.Where("`service` = ?", service).
			Where("`channel` = ?", channel).
			Where("`timestamp` = ?", timestamp).
			First(&record).Error
		if err == nil {
			id = record.ID
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
		id = record.ID
		return nil
	})
	return id, found, err
}

func (dao MysqlDao) GetReplyMessage(service string) (model.BotMessage, error) {
	msg := model.BotMessage{}
	if err := dao.db.Where("`service` = ?", service).First(&msg).Error; err != nil && !notFound(err) {
		return model.BotMessage{}, err
	}
	return msg, nil
}

func (dao MysqlDao) SetReplyMessage(msg model.BotMessage) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("`service` = ?", msg.Service).First(&model.BotMessage{}).Error
		if err == nil {
			return tx.Save(&msg).Error
		}

		if !notFound(err) {
			return err
		}

		return tx.Create(&msg).Error
	})
}

func (dao MysqlDao) GetSubscriber() ([]model.Subscriber, error) {
	subscribers := []model.Subscriber{}
	err := dao.db.Find(&subscribers).Error
	if err != nil {
		return nil, err
	}
	return subscribers, nil
}

func (dao MysqlDao) SetSubscriber(sub model.Subscriber) error {
	err := dao.db.Save(&sub).Error
	if err != nil {
		return err
	}
	return nil
}

func (dao MysqlDao) DeleteSubscriber(sub model.Subscriber) error {
	return dao.db.Where("`user_id` = ?", sub.UserID).Delete(&sub).Error
}
