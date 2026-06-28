package database

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var callerSkipPackages = []string{
	"github.com/alkuinvito/ai-assistant/pkg/database",
	"gorm.io/gorm",
}

type gormLogrus struct {
	logger    *logrus.Logger
	logLevel  gormlogger.LogLevel
	threshold time.Duration
}

func (l *gormLogrus) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.logLevel = level
	return l
}

func (l *gormLogrus) entry(ctx context.Context) *logrus.Entry {
	entry := l.logger.WithContext(ctx)
	if id := utils.GetRequestID(ctx); id != "" {
		entry = entry.WithField("req_id", id)
	}

	for i := 4; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		name := fn.Name()
		skip := false
		for _, pkg := range callerSkipPackages {
			if strings.HasPrefix(name, pkg) {
				skip = true
				break
			}
		}
		if !skip {
			entry = entry.WithField("caller", path.Base(name))
			entry = entry.WithField("trace", fmt.Sprintf("%s:%d", path.Base(file), line))
			break
		}
	}

	return entry
}

func (l *gormLogrus) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.entry(ctx).Infof(msg, data...)
	}
}

func (l *gormLogrus) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.entry(ctx).Warnf(msg, data...)
	}
}

func (l *gormLogrus) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.entry(ctx).Errorf(msg, data...)
	}
}

func (l *gormLogrus) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && l.logLevel >= gormlogger.Error {
		l.entry(ctx).WithFields(logrus.Fields{
			"elapsed": fmt.Sprintf("%.2fms", float64(elapsed.Microseconds())/1000),
		}).WithError(err).Errorf("[rows: %d] %s", rows, sql)
	} else if elapsed > l.threshold && l.logLevel >= gormlogger.Warn {
		l.entry(ctx).WithFields(logrus.Fields{
			"elapsed":   fmt.Sprintf("%.2fms", float64(elapsed.Microseconds())/1000),
			"threshold": l.threshold,
		}).Warnf("[rows: %d] %s", rows, sql)
	} else if l.logLevel >= gormlogger.Info {
		l.entry(ctx).WithFields(logrus.Fields{
			"elapsed": fmt.Sprintf("%.2fms", float64(elapsed.Microseconds())/1000),
		}).Debugf("[rows: %d] %s", rows, sql)
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
