package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/jsonx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const migrationVersion = "20260717_social_req"
const migrationDataVersion = "20260717_soc_data"
const receiptResultFixVersion = "20260718_receipt_result_fix"
const groupHandleMsgVersion = "20260718_group_handle_msg"
const invitationStatusSwapVersion = "20260719_invitation_status_swap"
const invitationReceiptCanonicalVersion = "20260719_invitation_receipt_canonical"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("socialmigration", flag.ContinueOnError)
	flags.SetOutput(stderr)
	driver := flags.String("driver", env("SOCIAL_MIGRATION_DRIVER", "mysql"), "database driver: sqlite, mysql, or postgres")
	dsn := flags.String("dsn", env("SOCIAL_MIGRATION_DSN", ""), "database DSN; prefer SOCIAL_MIGRATION_DSN")
	timeout := flags.Duration("timeout", 5*time.Minute, "migration timeout")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *dsn == "" {
		fmt.Fprintln(stderr, "database DSN is required")
		return 1
	}

	db, closeDB, err := openDatabase(*driver, *dsn)
	if err != nil {
		fmt.Fprintf(stderr, "open database: %v\n", err)
		return 1
	}
	defer func() {
		if err := closeDB(); err != nil {
			fmt.Fprintf(stderr, "close database: %v\n", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report, err := Migrate(ctx, db, normalizeDriver(*driver), time.Now().UTC())
	data, marshalErr := jsonx.Marshal(report)
	if marshalErr != nil {
		fmt.Fprintf(stderr, "encode report: %v\n", marshalErr)
		return 1
	}
	if _, writeErr := fmt.Fprintln(stdout, string(data)); writeErr != nil {
		fmt.Fprintf(stderr, "write report: %v\n", writeErr)
		return 1
	}
	if err != nil {
		fmt.Fprintf(stderr, "migration failed: %v\n", err)
		return 1
	}
	return 0
}

func openDatabase(driver, dsn string) (*gorm.DB, func() error, error) {
	var dialector gorm.Dialector
	switch normalizeDriver(driver) {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return nil, nil, fmt.Errorf("unsupported driver %q", driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		closeErr := sqlDB.Close()
		if closeErr != nil {
			return nil, nil, fmt.Errorf("ping database: %w; close database: %v", err, closeErr)
		}
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}
	return db, sqlDB.Close, nil
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgresql", "pg":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
