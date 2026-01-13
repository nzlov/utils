package gormlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nzlov/utils/otel"
	glogger "gorm.io/gorm/logger"
)

func init() {
	otel.SkipStr("gorm.io")
}

type GormLogger struct {
	LogLevel                  glogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
}

func Default() *GormLogger {
	return &GormLogger{
		LogLevel:                  glogger.Warn,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      true,
	}
}

// LogMode log mode
func (l *GormLogger) LogMode(level glogger.LogLevel) glogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= glogger.Info {
		otel.Info(ctx, fmt.Sprintf(msg, data...))
	}
}

// Warn print warn messages
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= glogger.Warn {
		otel.Warn(ctx, fmt.Sprintf(msg, data...))
	}
}

// Error print error messages
func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= glogger.Error {
		otel.Error(ctx, fmt.Sprintf(msg, data...))
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= glogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glogger.Error && (!errors.Is(err, glogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			otel.Error(ctx, fmt.Sprintf("[%.3fms][rows:-]%s Err:%s", float64(elapsed.Nanoseconds())/1e6, sql, err))
		} else {
			otel.Error(ctx, fmt.Sprintf("[%.3fms][rows:%v]%s Err:%s", float64(elapsed.Nanoseconds())/1e6, rows, sql, err))
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= glogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			otel.Warn(ctx, fmt.Sprintf("[%.3fms][rows:-]%s Warn:%s", float64(elapsed.Nanoseconds())/1e6, sql, slowLog))
		} else {
			otel.Warn(ctx, fmt.Sprintf("[%.3fms][rows:%v]%s Warn:%s", float64(elapsed.Nanoseconds())/1e6, rows, sql, slowLog))
		}
	case l.LogLevel == glogger.Info:
		sql, rows := fc()
		if rows == -1 {
			otel.Info(ctx, fmt.Sprintf("[%.3fms][rows:-]%s", float64(elapsed.Nanoseconds())/1e6, sql))
		} else {
			otel.Info(ctx, fmt.Sprintf("[%.3fms][rows:%v]%s", float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	}
}
