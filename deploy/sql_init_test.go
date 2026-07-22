package main

import (
	"fmt"
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

func TestFreshAutoMigrateRecordsReliabilityVersionAndReruns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/fresh.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	migrate := func(db *gorm.DB) error {
		return db.AutoMigrate(&objects.Friend{})
	}
	require.NoError(t, runAutoMigrateAndRecordFreshSchema(db, migrate))

	var count int64
	require.NoError(t, db.Model(&MigrationRecord{}).Where("version IN ?", migrationRecordVersions(freshSocialMigrationRecords)).Count(&count).Error)
	require.Equal(t, int64(len(freshSocialMigrationRecords)), count)
	require.NoError(t, runAutoMigrateAndRecordFreshSchema(db, migrate))
	require.NoError(t, db.Model(&MigrationRecord{}).Where("version IN ?", migrationRecordVersions(freshSocialMigrationRecords)).Count(&count).Error)
	require.Equal(t, int64(len(freshSocialMigrationRecords)), count)
}

func migrationRecordVersions(records []MigrationRecord) []string {
	versions := make([]string, 0, len(records))
	for _, record := range records {
		versions = append(versions, record.Version)
	}
	return versions
}

func TestLegacyAutoMigrateDoesNotRecordReliabilityVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/legacy.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`).Error)
	require.ErrorContains(t, runAutoMigrateAndRecordFreshSchema(db, func(*gorm.DB) error { return nil }), "deploy/socialmigration")
	require.False(t, db.Migrator().HasTable(&MigrationRecord{}))
}

func TestFreshAutoMigrateResumesAfterPartialDDL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/partial.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	injected := fmt.Errorf("injected migration failure")
	err = runAutoMigrateAndRecordFreshSchema(db, func(db *gorm.DB) error {
		require.NoError(t, db.AutoMigrate(&objects.Friend{}))
		return injected
	})
	require.ErrorIs(t, err, injected)

	require.NoError(t, runAutoMigrateAndRecordFreshSchema(db, func(db *gorm.DB) error {
		return db.AutoMigrate(&objects.Friend{})
	}))
	var bootstrap, canonical int64
	require.NoError(t, db.Model(&MigrationRecord{}).Where("version = ?", freshSocialBootstrapVersion).Count(&bootstrap).Error)
	require.NoError(t, db.Model(&MigrationRecord{}).Where("version = ?", socialRequestReliabilityMigrationVersion).Count(&canonical).Error)
	require.Zero(t, bootstrap)
	require.Equal(t, int64(1), canonical)
}

func TestFreshAutoMigrateRejectsPartialSchemaWithData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/partial-data.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	injected := fmt.Errorf("injected migration failure")
	require.ErrorIs(t, runAutoMigrateAndRecordFreshSchema(db, func(db *gorm.DB) error {
		require.NoError(t, db.AutoMigrate(&objects.Friend{}))
		return injected
	}), injected)
	require.NoError(t, db.Create(&objects.Friend{UserID: 1, FriendUID: 2}).Error)

	err = runAutoMigrateAndRecordFreshSchema(db, func(*gorm.DB) error { return nil })
	require.ErrorContains(t, err, "deploy/socialmigration")
	var bootstrap int64
	require.NoError(t, db.Model(&MigrationRecord{}).Where("version = ?", freshSocialBootstrapVersion).Count(&bootstrap).Error)
	require.Zero(t, bootstrap)
}
