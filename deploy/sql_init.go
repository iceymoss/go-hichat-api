package main

import (
	"fmt"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var migrationLogger *zap.Logger

func init() {
	var err error
	migrationLogger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize migration logger: %v", err))
	}
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"uniqueIndex;not null;size:20"`
	Description string    `gorm:"not null;size:255"`
	AppliedAt   time.Time `gorm:"not null"`
}

// TableName 指定迁移记录表名
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// InitMigrationTable 初始化迁移表
func InitMigrationTable(dbName string) error {
	mysqlConn := db.GetMysqlConn(dbName)

	// 使用 GORM AutoMigrate 创建迁移记录表
	err := mysqlConn.AutoMigrate(&MigrationRecord{})
	if err != nil {
		migrationLogger.Error("Failed to create migration table", zap.Error(err))
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	migrationLogger.Info("Migration table initialized successfully")
	return nil
}

// RunAutoMigrate 执行 GORM AutoMigrate 迁移
// 根据 pkg/db/objects 下的结构体自动创建/更新表结构
func RunAutoMigrate(dbName string) error {
	mysqlConn := db.GetMysqlConn(dbName)

	// 初始化迁移表
	if err := InitMigrationTable(dbName); err != nil {
		return err
	}

	// 定义所有需要迁移的表结构
	tables := []interface{}{
		// 用户模块
		&objects.User{},

		// 动态模块
		&objects.Trend{},
		&objects.TrendAgree{},
		&objects.TrendDiscuss{},

		// 社交模块
		&objects.Friend{},
		&objects.FriendRequest{},
		&objects.Group{},
		&objects.GroupMember{},
		&objects.GroupRequest{},

		// 可以在这里添加更多表结构
	}

	// 执行 AutoMigrate
	for _, table := range tables {
		tableName := getTableName(table)
		migrationLogger.Info("Migrating table", zap.String("table", tableName))

		if err := mysqlConn.AutoMigrate(table); err != nil {
			migrationLogger.Error("Failed to migrate table",
				zap.String("table", tableName),
				zap.Error(err))
			return fmt.Errorf("failed to migrate table %s: %w", tableName, err)
		}

		migrationLogger.Info("Table migrated successfully", zap.String("table", tableName))
	}

	migrationLogger.Info("All tables migrated successfully")
	return nil
}

// RunAutoMigrateWithVersion 执行带版本记录的 AutoMigrate
// 记录迁移版本，避免重复执行
func RunAutoMigrateWithVersion(dbName string, version string, description string) error {
	mysqlConn := db.GetMysqlConn(dbName)

	// 初始化迁移表
	if err := InitMigrationTable(dbName); err != nil {
		return err
	}

	// 检查迁移是否已执行
	var count int64
	mysqlConn.Model(&MigrationRecord{}).
		Where("version = ?", version).
		Count(&count)

	if count > 0 {
		migrationLogger.Info("Migration already applied, skipping",
			zap.String("version", version),
			zap.String("description", description))
		return nil
	}

	// 开始事务
	tx := mysqlConn.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			migrationLogger.Error("Migration panicked, rolled back",
				zap.String("version", version),
				zap.Any("panic", r))
		}
	}()

	// 在事务中执行 AutoMigrate
	if err := RunAutoMigrateInTx(tx); err != nil {
		tx.Rollback()
		migrationLogger.Error("Migration failed, rolled back",
			zap.String("version", version),
			zap.String("description", description),
			zap.Error(err))
		return fmt.Errorf("migration failed: %w", err)
	}

	// 记录迁移
	record := MigrationRecord{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now(),
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		migrationLogger.Error("Failed to record migration",
			zap.String("version", version),
			zap.Error(err))
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		migrationLogger.Error("Failed to commit migration",
			zap.String("version", version),
			zap.Error(err))
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	migrationLogger.Info("Migration applied successfully",
		zap.String("version", version),
		zap.String("description", description))

	return nil
}

// RunAutoMigrateInTx 在事务中执行 AutoMigrate
func RunAutoMigrateInTx(tx *gorm.DB) error {
	// 定义所有需要迁移的表结构
	tables := []interface{}{
		// 用户模块
		&objects.User{},

		// 动态模块
		&objects.Trend{},
		&objects.TrendAgree{},
		&objects.TrendDiscuss{},

		// 社交模块
		&objects.Friend{},
		&objects.FriendRequest{},
		&objects.Group{},
		&objects.GroupMember{},
		&objects.GroupRequest{},

		// 可以在这里添加更多表结构
	}

	// 执行 AutoMigrate
	for _, table := range tables {
		tableName := getTableName(table)
		if err := tx.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to migrate table %s: %w", tableName, err)
		}
	}

	return nil
}

// GetAppliedMigrations 获取已执行的迁移列表
func GetAppliedMigrations(dbName string) ([]MigrationRecord, error) {
	mysqlConn := db.GetMysqlConn(dbName)

	var records []MigrationRecord
	err := mysqlConn.Model(&MigrationRecord{}).
		Order("version ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	return records, nil
}

// getTableName 获取表名
func getTableName(table interface{}) string {
	if tn, ok := table.(interface{ TableName() string }); ok {
		return tn.TableName()
	}
	return fmt.Sprintf("%T", table)
}

// RunAllMigrations 执行所有迁移
// 在应用启动时调用此函数
func RunAllMigrations() error {
	dbName := db.MYSQL_DB_HICHAT2

	// 执行 AutoMigrate（不带版本记录，每次都会执行）
	// 注意：GORM AutoMigrate 是幂等的，只会添加缺失的字段，不会删除字段
	if err := RunAutoMigrate(dbName); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	return nil
}

// RunAllMigrationsWithVersion 执行所有迁移（带版本记录）
// 使用版本号控制，避免重复执行
func RunAllMigrationsWithVersion(version string, description string) error {
	dbName := db.MYSQL_DB_HICHAT2

	// 执行带版本记录的迁移
	if err := RunAutoMigrateWithVersion(dbName, version, description); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func main() {
	pkcCfg.InitConfig("local", "", "config")
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()
	err := RunAutoMigrateInTx(tx)
	if err != nil {
		panic(err)
	}
	fmt.Println("迁移成功")
}
