package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	driver := flag.String("driver", "mysql", "sqlite, mysql, or postgres")
	dsn := flag.String("dsn", "", "database DSN")
	idsText := flag.String("ids", "", "comma-separated dead outbox IDs")
	flag.Parse()
	if *dsn == "" || *idsText == "" {
		fmt.Fprintln(os.Stderr, "dsn and ids are required")
		os.Exit(2)
	}
	ids := make([]uint64, 0)
	for _, part := range strings.Split(*idsText, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err != nil || id == 0 {
			fmt.Fprintln(os.Stderr, "invalid id")
			os.Exit(2)
		}
		ids = append(ids, id)
	}
	var dialector gorm.Dialector
	switch *driver {
	case "sqlite":
		dialector = sqlite.Open(*dsn)
	case "postgres":
		dialector = postgres.Open(*dsn)
	default:
		dialector = mysql.Open(*dsn)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result := db.WithContext(context.Background()).Model(&objects.SocialNotificationOutbox{}).Where("id IN ? AND status = ?", ids, 2).Updates(map[string]any{"status": 0, "attempts": 0, "next_retry_at": nil, "last_error": "", "sent_at": nil})
	if result.Error != nil {
		fmt.Fprintln(os.Stderr, result.Error)
		os.Exit(1)
	}
	fmt.Printf("replayed=%d\n", result.RowsAffected)
}
