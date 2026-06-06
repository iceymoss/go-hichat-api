package systemconfig

import (
	"context"
	"strconv"
	"strings"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
)

func GetString(ctx context.Context, key string, defaultVal string) string {
	var row objects.SystemSetting
	err := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2)).
		Where("key_name = ?", key).
		First(&row).Error
	if err != nil || strings.TrimSpace(row.Value) == "" {
		return defaultVal
	}
	return row.Value
}

func GetBool(ctx context.Context, key string, defaultVal bool) bool {
	val := strings.ToLower(strings.TrimSpace(GetString(ctx, key, "")))
	if val == "" {
		return defaultVal
	}
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func GetInt(ctx context.Context, key string, defaultVal int) int {
	val := strings.TrimSpace(GetString(ctx, key, ""))
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func GetStringList(ctx context.Context, key string, defaultVal []string) []string {
	val := strings.TrimSpace(GetString(ctx, key, ""))
	if val == "" {
		return defaultVal
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return defaultVal
	}
	return out
}
