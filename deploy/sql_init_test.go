package main

import (
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSocialReliabilityAutoMigrateGate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T, *gorm.DB)
		wantError bool
	}{
		{name: "fresh database creates complete schema"},
		{name: "legacy database is blocked", wantError: true, setup: func(t *testing.T, db *gorm.DB) {
			require.NoError(t, db.Exec(`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`).Error)
		}},
		{name: "completed dedicated migration permits auto migrate", setup: func(t *testing.T, db *gorm.DB) {
			require.NoError(t, db.Exec(`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`).Error)
			require.NoError(t, db.AutoMigrate(&MigrationRecord{}))
			require.NoError(t, db.Create(&MigrationRecord{Version: socialRequestReliabilityMigrationVersion, Description: "test", AppliedAt: time.Now()}).Error)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(t.TempDir()+"/gate.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			if test.setup != nil {
				test.setup(t, db)
			}
			err = requireSocialReliabilityMigration(db)
			if test.wantError {
				require.ErrorContains(t, err, "deploy/socialmigration")
				require.False(t, db.Migrator().HasIndex(&objects.Friend{}, "uk_friends_user_friend"))
				return
			}
			require.NoError(t, err)
			if !db.Migrator().HasTable(&objects.Friend{}) {
				require.NoError(t, db.AutoMigrate(&objects.Friend{}, &objects.GroupInvitation{}, &objects.SocialRequestReceipt{}, &objects.SocialNotificationOutbox{}))
				require.True(t, db.Migrator().HasIndex(&objects.Friend{}, "uk_friends_user_friend"))
				require.True(t, db.Migrator().HasTable(&objects.GroupInvitation{}))
			}
		})
	}
}

func TestMigrationTablesContainReliabilityObjects(t *testing.T) {
	wanted := map[string]bool{
		"group_invitations":          false,
		"social_request_receipts":    false,
		"social_notification_outbox": false,
	}
	for _, table := range migrationTables() {
		if _, ok := wanted[getTableName(table)]; ok {
			wanted[getTableName(table)] = true
		}
	}
	for table, present := range wanted {
		require.True(t, present, "missing %s from migrationTables", table)
	}
}
