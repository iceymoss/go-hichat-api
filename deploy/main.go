package main

import (
	"fmt"
	"time"

	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/gorm"
)

const socialRequestReliabilityMigrationVersion = "20260717_social_req"

type migrationRecord struct {
	ID          uint64    `gorm:"primaryKey"`
	Version     string    `gorm:"size:64;not null;uniqueIndex"`
	Description string    `gorm:"size:255;not null"`
	AppliedAt   time.Time `gorm:"not null"`
}

func (migrationRecord) TableName() string { return "schema_migrations" }

func main() {
	pkcCfg.InitConfig("local", "config")
	if err := migrate(db.GetMysqlConn(db.MYSQL_DB_HICHAT2)); err != nil {
		panic(err)
	}
	fmt.Println("数据库迁移完成")
}

func migrate(conn *gorm.DB) error {
	if err := requireSocialReliabilityMigration(conn); err != nil {
		return err
	}
	for _, table := range tables() {
		if err := conn.AutoMigrate(table); err != nil {
			return fmt.Errorf("migrate %T: %w", table, err)
		}
	}
	return nil
}

func requireSocialReliabilityMigration(conn *gorm.DB) error {
	if !hasCoreSocialTables(conn) {
		return nil
	}
	if !conn.Migrator().HasTable(&migrationRecord{}) {
		return fmt.Errorf("legacy social schema requires the explicit social reliability migration %s before AutoMigrate", socialRequestReliabilityMigrationVersion)
	}
	var count int64
	if err := conn.Model(&migrationRecord{}).Where("version = ?", socialRequestReliabilityMigrationVersion).Count(&count).Error; err != nil {
		return fmt.Errorf("check social reliability migration: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("legacy social schema requires the explicit social reliability migration %s before AutoMigrate", socialRequestReliabilityMigrationVersion)
	}
	return nil
}

func hasCoreSocialTables(conn *gorm.DB) bool {
	for _, table := range []any{&objects.Friend{}, &objects.FriendRequest{}, &objects.Group{}, &objects.GroupMember{}, &objects.GroupRequest{}} {
		if conn.Migrator().HasTable(table) {
			return true
		}
	}
	return false
}

func tables() []any {
	return []any{
		&objects.User{},
		&objects.UserSettings{},
		&objects.Favorite{},
		&objects.UserEmoji{},
		&objects.Trend{},
		&objects.TrendAgree{},
		&objects.TrendDiscuss{},
		&objects.TrendDraft{},
		&objects.TrendMessage{},
		&objects.Friend{},
		&objects.FriendRequest{},
		&objects.FriendReport{},
		&objects.Group{},
		&objects.GroupMember{},
		&objects.GroupRequest{},
		&objects.GroupInvitation{},
		&objects.SocialRequestReceipt{},
		&objects.SocialNotificationOutbox{},
		&objects.GroupInviteLink{},
		&objects.GroupMemberSetting{},
		&objects.GroupAnnouncement{},
		&objects.RelationOutbox{},
		&objects.Notification{},
		&objects.NotificationReadIntent{},
		&objects.SystemSetting{},
	}
}
