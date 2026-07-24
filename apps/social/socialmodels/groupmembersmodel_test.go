package socialmodels

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGroupMembersListAllowsNullInvitationMetadata(t *testing.T) {
	dsn := os.Getenv("SOCIAL_GROUP_MEMBERS_TEST_MYSQL_DSN")
	if dsn == "" || os.Getenv("HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS") != "1" {
		t.Skip("set SOCIAL_GROUP_MEMBERS_TEST_MYSQL_DSN and HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS=1 to run")
	}
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	var databaseName string
	require.NoError(t, database.Raw("SELECT DATABASE()").Scan(&databaseName).Error)
	matched, err := regexp.MatchString(`^hichat_[a-z0-9_]*_test$`, strings.ToLower(databaseName))
	require.NoError(t, err)
	require.True(t, matched, "test requires a dedicated hichat_*_test database")

	require.NoError(t, database.Exec("DROP TABLE IF EXISTS group_members").Error)
	require.NoError(t, database.Exec(`CREATE TABLE group_members (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		group_id BIGINT UNSIGNED NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		role_level TINYINT NOT NULL,
		join_time TIMESTAMP NULL,
		join_source TINYINT NULL,
		inviter_uid BIGINT UNSIGNED NULL,
		operator_uid BIGINT UNSIGNED NULL,
		group_nickname VARCHAR(64) NOT NULL DEFAULT '',
		group_remark VARCHAR(255) NOT NULL DEFAULT ''
	)`).Error)
	require.NoError(t, database.Exec("INSERT INTO group_members (group_id, user_id, role_level) VALUES (1, 2, 0)").Error)

	cacheConf := cache.CacheConf{{RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}, Weight: 100}}
	model := &defaultGroupMembersModel{CachedConn: sqlc.NewConn(sqlx.NewMysql(dsn), cacheConf), table: "`group_members`", mysqlConn: database}
	members, err := model.ListByGroupId(context.Background(), "1")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.False(t, members[0].JoinTime.Valid)
	require.False(t, members[0].JoinSource.Valid)
	require.False(t, members[0].InviterUid.Valid)
	require.False(t, members[0].OperatorUid.Valid)
}
