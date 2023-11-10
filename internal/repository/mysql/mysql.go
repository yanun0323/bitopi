package mysql

import (
	"bitopi/internal/model"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_driverKey = struct{}{}
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
		&model.BotSetting{},
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

func (dao MysqlDao) Tx(ctx context.Context, fn func(context.Context) error) error {
	_, ok := ctx.Value(_driverKey).(*gorm.DB)
	if ok {
		return errors.New("multiple transaction")
	}
	return dao.GetDriver(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, _driverKey, tx)
		return fn(ctx)
	})
}

func (dao MysqlDao) GetDriver(ctx context.Context) (db *gorm.DB) {
	db, ok := ctx.Value(_driverKey).(*gorm.DB)
	if ok && db != nil {
		return db
	}
	return dao.db
}

func (dao MysqlDao) GetMember(ctx context.Context, service string, userID string) (model.Member, error) {
	var member model.Member
	err := dao.GetDriver(ctx).
		Where("`service` = ?", service).
		Where("`user_id` = ?", userID).
		First(&member).Error
	if err != nil {
		return model.Member{}, err
	}
	return member, nil
}

func (dao MysqlDao) UpdateMember(ctx context.Context, member model.Member) error {
	var id uint64
	err := dao.GetDriver(ctx).
		Model(&model.Member{}).
		Select("`id`").
		Where("`service` = ?", member.Service).
		Where("`user_id` = ?", member.UserID).
		First(&id).Error
	if err != nil {
		return err
	}
	member.ID = id
	return dao.GetDriver(ctx).Save(&member).Error
}

func (dao MysqlDao) ListMembers(ctx context.Context, service string) ([]model.Member, error) {
	var members []model.Member
	err := dao.GetDriver(ctx).Where("`service` = ?", service).
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

func (dao MysqlDao) ResetMembers(txCtx context.Context, service string, member []model.Member) error {
	tx := dao.GetDriver(txCtx)

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
}

func (dao MysqlDao) ListAllMembers(ctx context.Context) ([]model.Member, error) {
	var members []model.Member
	err := dao.GetDriver(ctx).Find(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (dao MysqlDao) IsAdmin(ctx context.Context, service, userID string) (bool, error) {
	err := dao.GetDriver(ctx).
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

func (dao MysqlDao) ListAdmin(ctx context.Context, service string) ([]model.Admin, error) {
	var admins []model.Admin
	if err := dao.GetDriver(ctx).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (dao MysqlDao) AddAdmin(ctx context.Context, admin model.Admin) error {
	if admin.IsEmpty() {
		return errors.New(fmt.Sprintf("empty admin, %+v", admin))
	}
	return dao.GetDriver(ctx).Save(&admin).Error
}

func (dao MysqlDao) DeleteAdmin(ctx context.Context, service, userID string) error {
	err := dao.GetDriver(ctx).
		Where("`service` = ?", service).
		Where("`user_id` = ?", userID).
		Delete(&model.Admin{}).Error
	if err != nil && !notFound(err) {
		return err
	}
	return nil
}

func (dao MysqlDao) GetStartDate(ctx context.Context, service string) (time.Time, error) {
	elem := model.StartTime{}
	err := dao.GetDriver(ctx).Where("`service` = ?", service).
		First(&elem).Error
	if err != nil {
		return time.Time{}, err
	}

	return elem.StartTime, nil
}

func (dao MysqlDao) UpdateStartDate(txCtx context.Context, service string, t time.Time) error {
	tx := dao.GetDriver(txCtx)
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
}

func (dao MysqlDao) GetDutyDuration(ctx context.Context, service string) (time.Duration, error) {
	var setting model.BotSetting
	key := strings.ToLower(service) + ".duty.duration"
	if err := dao.GetDriver(ctx).Model(&setting).Where("`key` = ?", key).Limit(1).First(&setting).Error; err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(setting.Value)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func (dao MysqlDao) GetDutyMemberCountPerTime(ctx context.Context, service string) (int, error) {
	var setting model.BotSetting
	key := strings.ToLower(service) + ".duty.member.count.per.time"
	if err := dao.GetDriver(ctx).Model(&setting).Where("`key` = ?", key).Limit(1).First(&setting).Error; err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(setting.Value)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao MysqlDao) CountMentionRecord(ctx context.Context, service string) (int64, error) {
	var count int64
	if err := dao.GetDriver(ctx).Model(&model.MentionRecord{}).Where("`service` = ?", service).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (dao MysqlDao) GetMentionRecord(ctx context.Context, id uint64) (model.MentionRecord, error) {
	record := model.MentionRecord{}
	if err := dao.GetDriver(ctx).Where("`id`= ?", id).First(&record).Error; err != nil {
		return model.MentionRecord{}, err
	}
	return record, nil
}

func (dao MysqlDao) FindOrCreateMentionRecord(txCtx context.Context, service, channel, timestamp string) (uint64, bool, error) {
	found := false
	var id uint64
	tx := dao.GetDriver(txCtx)

	record := model.MentionRecord{}
	err := tx.Where("`service` = ?", service).
		Where("`channel` = ?", channel).
		Where("`timestamp` = ?", timestamp).
		First(&record).Error
	if err == nil {
		id = record.ID
		found = true
		return id, found, nil
	}

	if !notFound(err) {
		return id, found, errors.Wrap(err, "query")
	}

	record.Service = service
	record.Channel = channel
	record.Timestamp = timestamp
	record.CreateAtu = time.Now().Unix()
	if err := tx.Create(&record).Error; err != nil {
		return id, found, errors.Wrap(err, "create")
	}
	id = record.ID
	return id, found, nil
}

func (dao MysqlDao) GetReplyMessage(ctx context.Context, service string) (model.BotMessage, error) {
	msg := model.BotMessage{}
	if err := dao.GetDriver(ctx).Where("`service` = ?", service).First(&msg).Error; err != nil && !notFound(err) {
		return model.BotMessage{}, err
	}
	return msg, nil
}

func (dao MysqlDao) SetReplyMessage(txCtx context.Context, msg model.BotMessage) error {
	tx := dao.GetDriver(txCtx)
	var id uint64
	err := tx.Model(&model.BotMessage{}).Select("`id`").Where("`service` = ?", msg.Service).First(&id).Error
	if notFound(err) {
		return tx.Create(&msg).Error
	}

	if err != nil {
		return err
	}

	msg.ID = id
	return tx.Save(&msg).Error
}

func (dao MysqlDao) GetSubscriber(ctx context.Context) ([]model.Subscriber, error) {
	subscribers := []model.Subscriber{}
	err := dao.GetDriver(ctx).Find(&subscribers).Error
	if err != nil {
		return nil, err
	}
	return subscribers, nil
}

func (dao MysqlDao) SetSubscriber(ctx context.Context, sub model.Subscriber) error {
	err := dao.GetDriver(ctx).Save(&sub).Error
	if err != nil {
		return err
	}
	return nil
}

func (dao MysqlDao) DeleteSubscriber(ctx context.Context, sub model.Subscriber) error {
	return dao.GetDriver(ctx).Where("`user_id` = ?", sub.UserID).Delete(&sub).Error
}
