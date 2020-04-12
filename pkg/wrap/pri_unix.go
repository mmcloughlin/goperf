// +build !linux

package wrap

import "go.uber.org/zap"

func prioritizeactions(l *zap.Logger) []action {
	return []action{
		&niceaction{},
	}
}
