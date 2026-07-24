package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSocialReliabilityMigrationGate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T, *gorm.DB)
		wantError bool
	}{
		{name: "fresh database"},
		{name: "legacy database without marker", wantError: true, setup: func(t *testing.T, db *gorm.DB) {
			require.NoError(t, db.Exec(`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`).Error)
		}},
		{name: "explicit migration marker", setup: func(t *testing.T, db *gorm.DB) {
			require.NoError(t, db.Exec(`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`).Error)
			require.NoError(t, db.AutoMigrate(&migrationRecord{}))
			require.NoError(t, db.Create(&migrationRecord{Version: socialRequestReliabilityMigrationVersion, Description: "test", AppliedAt: time.Now()}).Error)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(t.TempDir()+"/migration.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			if tt.setup != nil {
				tt.setup(t, db)
			}
			err = requireSocialReliabilityMigration(db)
			if tt.wantError {
				require.ErrorContains(t, err, "explicit social reliability migration")
				return
			}
			require.NoError(t, err)
		})
	}
}
