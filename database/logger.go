package database

import (
	"context"
	"fmt"
	"time"

	appLogger "github.com/xtj/ai-argus/pkg/logger"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

const slowQueryThreshold = 200 * time.Millisecond

type zapGormLogger struct {
	level gormLogger.LogLevel
}

func newGormLogger(level string) gormLogger.Interface {
	logLevel := gormLogger.Warn
	switch level {
	case "silent":
		logLevel = gormLogger.Silent
	case "error":
		logLevel = gormLogger.Error
	case "info":
		logLevel = gormLogger.Info
	}
	return &zapGormLogger{level: logLevel}
}

func (l *zapGormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &zapGormLogger{level: level}
}

func (l *zapGormLogger) ParamsFilter(_ context.Context, sql string, _ ...any) (string, []any) {
	return sql, nil
}

func (l *zapGormLogger) Info(_ context.Context, msg string, data ...any) {
	if l.level >= gormLogger.Info {
		appLogger.L().Info(fmt.Sprintf(msg, data...))
	}
}

func (l *zapGormLogger) Warn(_ context.Context, msg string, data ...any) {
	if l.level >= gormLogger.Warn {
		appLogger.L().Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *zapGormLogger) Error(_ context.Context, msg string, data ...any) {
	if l.level >= gormLogger.Error {
		appLogger.L().Error(fmt.Sprintf(msg, data...))
	}
}

func (l *zapGormLogger) Trace(_ context.Context, begin time.Time, query func() (string, int64), err error) {
	if l.level == gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.level >= gormLogger.Error:
		sql, rows := query()
		appLogger.L().Error("gorm query failed", zap.Error(err), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case elapsed > slowQueryThreshold && l.level >= gormLogger.Warn:
		sql, rows := query()
		appLogger.L().Warn("gorm slow query", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.level >= gormLogger.Info:
		sql, rows := query()
		appLogger.L().Info("gorm query", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}
