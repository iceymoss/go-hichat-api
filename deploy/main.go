package main

import (
	"fmt"

	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/gorm"
)

func main() {
	pkcCfg.InitConfig("local", "config")
	if err := migrate(db.GetMysqlConn(db.MYSQL_DB_HICHAT2)); err != nil {
		panic(err)
	}
	fmt.Println("数据库迁移完成")
}

func migrate(conn *gorm.DB) error {
	for _, table := range tables() {
		if err := conn.AutoMigrate(table); err != nil {
			return fmt.Errorf("migrate %T: %w", table, err)
		}
	}
	return nil
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
