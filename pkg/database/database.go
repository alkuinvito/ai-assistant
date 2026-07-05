package database

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormlogger "gorm.io/gorm/logger"
)

type gormLogrus struct {
	logger    *logrus.Logger
	logLevel  logger.LogLevel
	threshold time.Duration
}

func (l *gormLogrus) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

func (l *gormLogrus) entry(ctx context.Context) *logrus.Entry {
	entry := l.logger.WithContext(ctx)
	if id := utils.GetRequestID(ctx); id != "" {
		entry = entry.WithField("req_id", id)
	}

	pc, file, line, ok := runtime.Caller(4)
	if !ok {
		return entry
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return entry
	}
	name := fn.Name()
	entry = entry.WithField("caller", path.Base(name))
	entry = entry.WithField("trace", fmt.Sprintf("%s:%d", path.Base(file), line))

	return entry
}

func (l *gormLogrus) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.entry(ctx).Infof(msg, data...)
	}
}

func (l *gormLogrus) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.entry(ctx).Warnf(msg, data...)
	}
}

func (l *gormLogrus) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.entry(ctx).Errorf(msg, data...)
	}
}

func (l *gormLogrus) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && l.logLevel >= logger.Error {
		l.entry(ctx).WithFields(logrus.Fields{
			"duration": elapsed.Microseconds(),
			"rows":     rows,
		}).WithError(err).Error(sql)
	} else if elapsed > l.threshold && l.logLevel >= logger.Warn {
		l.entry(ctx).WithFields(logrus.Fields{
			"duration": elapsed.Microseconds(),
			"rows":     rows,
		}).Warn(sql)
	} else if l.logLevel >= logger.Info {
		l.entry(ctx).WithFields(logrus.Fields{
			"duration": elapsed.Microseconds(),
			"rows":     rows,
		}).Info(sql)
	}
}

func NewDatabase(logger *logrus.Logger) (*gorm.DB, func(), error) {
	appEnv := os.Getenv("APP_ENV")

	var logLevel gormlogger.LogLevel
	if appEnv == "development" {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Warn
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("DATABASE_URL"),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		TranslateError: true,
		Logger: &gormLogrus{
			logger:    logger,
			logLevel:  logLevel,
			threshold: 200 * time.Millisecond,
		},
	})

	return db, cleanup(db, logger), err
}

func cleanup(db *gorm.DB, logger *logrus.Logger) func() {
	return func() {
		sqlDB, err := db.DB()
		if err != nil {
			logger.WithError(err).Fatal("failed to get sql db")
		}

		err = sqlDB.Close()
		if err != nil {
			logger.WithError(err).Fatal("failed to close sql db")
		}
	}
}
