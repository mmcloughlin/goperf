// Package lg provides logging utilities.
package lg

import (
	"time"

	"go.uber.org/zap"
)

func Scope(l *zap.Logger, name string, fields ...zap.Field) func() {
	t0 := time.Now()
	l.Debug("start "+name, fields...)
	return func() {
		l.Debug("finish "+name, zap.Duration("duration", time.Since(t0)))
	}
}
