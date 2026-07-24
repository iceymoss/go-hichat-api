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

const (
	exitSuccess         = 0
	exitFailure         = 1
	exitCleanupRequired = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func run(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	flags := flag.NewFlagSet("migrationaudit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	driver := flags.String("driver", env("MIGRATION_AUDIT_DRIVER", "mysql"), "database driver: mysql, postgres, or sqlite")
	dsn := flags.String("dsn", env("MIGRATION_AUDIT_DSN", ""), "database DSN; prefer MIGRATION_AUDIT_DSN to avoid shell history")
	timeout := flags.Duration("timeout", 30*time.Second, "audit timeout")
	if err := flags.Parse(args); err != nil {
		return exitFailure
	}
	if *dsn == "" {
		fmt.Fprintln(stderr, "database DSN is required")
		return exitFailure
	}

	db, closeDB, err := openDatabase(*driver, *dsn)
	if err != nil {
		fmt.Fprintf(stderr, "open database: %v\n", err)
		return exitFailure
	}
	defer func() {
		if err := closeDB(); err != nil {
			fmt.Fprintf(stderr, "close database: %v\n", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report, err := NewAuditor(db).Run(ctx, Database{
		Driver: normalizeDriver(*driver), Name: databaseName(normalizeDriver(*driver), *dsn),
	}, now().UTC().Format(time.RFC3339))
	if err != nil {
		fmt.Fprintf(stderr, "run audit: %v\n", err)
		return exitFailure
	}

	data, err := jsonx.Marshal(report)
	if err != nil {
		fmt.Fprintf(stderr, "encode report: %v\n", err)
		return exitFailure
	}
	if _, err := fmt.Fprintln(stdout, string(data)); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return exitFailure
	}
	if report.Summary.RequiresCleanup {
		return exitCleanupRequired
	}
	return exitSuccess
}

func openDatabase(driver, dsn string) (*gorm.DB, func() error, error) {
	var dialector gorm.Dialector
	switch normalizeDriver(driver) {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
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
