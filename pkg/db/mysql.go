package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	conf "github.com/iceymoss/go-hichat-api/pkg/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const MYSQL_DB_HICHAT2 = "hichat2"

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize logger: %v", err))
	}
}

var mysqlConn = make(map[string]*gorm.DB)
var mysqlMutex sync.RWMutex

func GetMysqlConn(db string) *gorm.DB {
	mysqlMutex.RLock()
	conn, ok := mysqlConn[db]
	mysqlMutex.RUnlock()
	if !ok {
		mysqlMutex.Lock()
		userName := conf.ServiceConf.DB.User
		userPwd := conf.ServiceConf.DB.Password
		host := conf.ServiceConf.DB.Host
		port := strconv.Itoa(conf.ServiceConf.DB.Port)
		envLogLevel := conf.ServiceConf.DB.LogLevel

		var gormlevel gormLogger.LogLevel
		switch envLogLevel {
		case "error":
			gormlevel = gormLogger.Error
		case "debug":
			gormlevel = gormLogger.Info
		case "info":
			gormlevel = gormLogger.Info
		case "warning":
			gormlevel = gormLogger.Warn
		case "fatal":
			gormlevel = gormLogger.Error
		case "panic":
			gormlevel = gormLogger.Error
		case "dpanic":
			gormlevel = gormLogger.Error
		default:
			gormlevel = gormLogger.Warn
		}
		dsn := userName + ":" + userPwd + "@tcp(" + host + ":" + port + ")/" + db + "?charset=utf8mb4&parseTime=True&loc=Local"
		dbConn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: &CustomMySqlLogger{
				Logger: logger,
				Config: gormLogger.Config{
					LogLevel:                  gormlevel,
					Colorful:                  false,
					IgnoreRecordNotFoundError: true,
					SlowThreshold:             500 * time.Millisecond,
				},
			},
		})

		pool, pollErr := dbConn.DB()
		if pollErr != nil {
			logger.Error(pollErr.Error())
		} else {
			pool.SetMaxOpenConns(30)
			pool.SetMaxIdleConns(15)
		}

		logger.Debug("db conntion dsn " + dsn)
		if err != nil {
			logger.Error(err.Error())
		} else {
			if conf.ServiceConf.DB.LogLevel == "debug" {
				mysqlConn[db] = dbConn.Debug()
			} else {
				mysqlConn[db] = dbConn
			}
		}
		conn = dbConn
		mysqlMutex.Unlock()
	}

	return conn
}

type CustomMySqlLogger struct {
	Logger *zap.Logger
	Config gormLogger.Config
}

func (l *CustomMySqlLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {

	newlogger := *l

	newlogger.Config.LogLevel = level

	return &newlogger

}

func (l CustomMySqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {

	defer l.Logger.Sync()

	l.Logger.Error(fmt.Sprintf("%s", append([]interface{}{msg}, data...)...),

		zap.String("source", utils.FileWithLineNum()),

		zap.String("agg_type", "gorm"),
	)

}

func (l CustomMySqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	defer l.Logger.Sync()
	l.Logger.Error(fmt.Sprintf("%s", append([]interface{}{msg}, data...)...),
		zap.String("source", utils.FileWithLineNum()),
		zap.String("agg_type", "gorm"),
	)
}

// Error print error messages
func (l CustomMySqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	defer l.Logger.Sync()
	l.Logger.Error(fmt.Sprintf("%s", append([]interface{}{msg}, data...)...),
		zap.String("source", utils.FileWithLineNum()),
		zap.Any("source", utils.FileWithLineNum()),
		zap.String("agg_type", "gorm"),
	)
}

// Trace 开启事务处理
func (l CustomMySqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	defer l.Logger.Sync()
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Config.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError):
		sql, rows := fc()
		l.Logger.Error(err.Error(),
			zap.String("source", utils.FileWithLineNum()),
			zap.Float64("query_time", float64(elapsed.Nanoseconds())/1e6),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
			zap.String("agg_type", "gorm"),
		)

	case elapsed > l.Config.SlowThreshold && l.Config.SlowThreshold != 0 && l.Config.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.Config.SlowThreshold)
		l.Logger.Warn(slowLog,
			zap.String("source", utils.FileWithLineNum()),
			zap.Float64("query_time", float64(elapsed.Nanoseconds())/1e6),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
			zap.String("agg_type", "gorm"),
		)

	case l.Config.LogLevel == gormLogger.Info:
		sql, rows := fc()
		l.Logger.Warn("sql log",
			zap.String("source", utils.FileWithLineNum()),
			zap.Float64("query_time", float64(elapsed.Nanoseconds())/1e6),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
			zap.String("agg_type", "gorm"),
		)
	}
}
