// +build !linux

package wrap

import "github.com/mmcloughlin/cb/pkg/lg"

func prioritizeactions(l lg.Logger) []action {
	return []action{
		&niceaction{},
	}
}
