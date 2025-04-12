package wuid

import (
	"database/sql"
	"fmt"
	"github.com/edwingeng/wuid/mysql/wuid"
)

var w *wuid.WUID

func Init(dsn string) {

	newDB := func() (*sql.DB, bool, error) {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, false, err
		}
		return db, true, nil
	}

	w = wuid.NewWUID("default", nil)
	_ = w.LoadH28FromMysql(newDB, "wuid")
}

func GenUid(dsn string) string {
	if w == nil {
		Init(dsn)
	}

	return fmt.Sprintf("%#016x", w.Next())
}
