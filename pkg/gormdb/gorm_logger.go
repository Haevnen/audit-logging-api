package gormdb

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	gormLogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	LogLevel gormLogger.LogLevel
	base     *logrus.Logger
}

func NewGormLogger(level gormLogger.LogLevel, base *logrus.Logger) *GormLogger {
	return &GormLogger{LogLevel: level, base: base}
}

func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		l.base.Infof(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		l.base.Warnf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		l.base.Errorf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= gormLogger.Error:
		l.base.WithError(err).Errorf("%s [%s] rows:%d", sql, elapsed, rows)
	case l.LogLevel >= gormLogger.Info:
		l.base.Infof("%s [%s] rows:%d", sql, elapsed, rows)
	}
}
