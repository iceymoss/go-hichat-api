package main

import (
	"fmt"

	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
)

// trend 表维护工具（本分支联调用，可在跑完后删除）：
//  1. 给 trend 表补 scope 列(ADD COLUMN，幂等)，绕开 users.uk_email 的历史脏数据；
//  2. 用真实有效点赞(trend_agree.state=1)重算 agree_count，修复历史被减成负数/下溢的计数。
func main() {
	pkcCfg.InitConfig("local", "", "config")
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)

	// 1. scope 列
	if err := mysqlConn.AutoMigrate(&objects.Trend{}); err != nil {
		panic(err)
	}
	fmt.Println("trend.scope 迁移完成")

	// 2. 重算 agree_count（关联子查询，MySQL/PostgreSQL/SQLite 通用）
	res := mysqlConn.Exec(
		"UPDATE trend SET agree_count = (SELECT COUNT(*) FROM trend_agree WHERE trend_agree.trend_id = trend.id AND trend_agree.state = 1)",
	)
	if res.Error != nil {
		panic(res.Error)
	}
	fmt.Printf("agree_count 重算完成，影响 %d 行\n", res.RowsAffected)
}
