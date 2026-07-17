package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/jsonx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAuditorRun(t *testing.T) {
	db := newAuditTestDB(t)
	require.NoError(t, seedAuditTestDB(db))

	report, err := NewAuditor(db).Run(context.Background(), Database{Driver: "sqlite", Name: "test"}, "2026-07-17T00:00:00Z")
	require.NoError(t, err)
	require.True(t, report.Summary.RequiresCleanup)
	require.Equal(t, 7, report.Summary.FindingGroups)
	require.Equal(t, 6, report.Summary.AffectedUsers)

	require.Equal(t, []uint64{10, 11}, report.Findings.DuplicateFriendRequests[0].RecordIDs)
	require.Equal(t, uint64(11), report.Findings.DuplicateFriendRequests[0].KeepID)
	require.Equal(t, []int64{20, 21}, report.Findings.DuplicateGroupRequests[0].RecordIDs)
	require.Equal(t, int64(21), report.Findings.DuplicateGroupRequests[0].KeepID)
	require.Len(t, report.Findings.LegacyGroupInvitations, 2)
	require.Equal(t, "recoverable_member_invitation", report.Findings.LegacyGroupInvitations[0].Classification)
	require.Equal(t, "invite_link_not_user_invitation", report.Findings.LegacyGroupInvitations[1].Classification)
	require.Equal(t, []uint64{30, 31}, report.Findings.DuplicateFriends[0].RecordIDs)
	require.Equal(t, uint64(30), report.Findings.DuplicateFriends[0].KeepID)
	require.Len(t, report.Findings.OneWayFriends, 2)
	require.True(t, report.Findings.GroupMemberUniqueIndex.Valid)
}

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name string
		seed bool
		want int
	}{
		{name: "clean", want: exitSuccess},
		{name: "cleanup required", seed: true, want: exitCleanupRequired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/audit.db"
			db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			require.NoError(t, createAuditSchema(db))
			if test.seed {
				require.NoError(t, seedAuditTestDB(db))
			}
			sqlDB, err := db.DB()
			require.NoError(t, err)
			require.NoError(t, sqlDB.Close())

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run([]string{"-driver", "sqlite", "-dsn", path}, &stdout, &stderr, func() time.Time {
				return time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
			})
			require.Equal(t, test.want, code, stderr.String())
			var report Report
			require.NoError(t, jsonx.Unmarshal(stdout.Bytes(), &report))
			require.Equal(t, test.seed, report.Summary.RequiresCleanup)
		})
	}
}

func TestRunRejectsUnsupportedDriver(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-driver", "oracle", "-dsn", "ignored"}, &stdout, &stderr, time.Now)
	require.Equal(t, exitFailure, code)
	require.Contains(t, stderr.String(), "unsupported driver")
}

func newAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/audit.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, createAuditSchema(db))
	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			require.NoError(t, sqlDB.Close())
		}
	})
	return db
}

func createAuditSchema(db *gorm.DB) error {
	statements := []string{
		`CREATE TABLE friend_requests (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, req_uid INTEGER NOT NULL, handle_result INTEGER, status INTEGER)`,
		`CREATE TABLE group_requests (id INTEGER PRIMARY KEY, req_id TEXT NOT NULL, group_id TEXT NOT NULL, join_source INTEGER, inviter_user_id TEXT, handle_result INTEGER)`,
		`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL)`,
		`CREATE TABLE group_members (id INTEGER PRIMARY KEY, group_id INTEGER NOT NULL, user_id INTEGER NOT NULL)`,
		`CREATE UNIQUE INDEX uk_member ON group_members(group_id, user_id)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAuditTestDB(db *gorm.DB) error {
	statements := []string{
		`INSERT INTO friend_requests(id,user_id,req_uid,handle_result,status) VALUES (10,1,2,0,1),(11,1,2,0,1),(12,1,3,1,1)`,
		`INSERT INTO group_requests(id,req_id,group_id,join_source,inviter_user_id,handle_result) VALUES (20,'1','100',1,NULL,0),(21,'1','100',1,NULL,0),(22,'4','100',2,'3',0),(23,'5','100',3,'3',1)`,
		`INSERT INTO friends(id,user_id,friend_uid) VALUES (30,1,2),(31,1,2),(32,2,1),(33,3,4),(34,5,6)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
