package lg

import "time"

type Interface interface {
	Printf(format string, v ...interface{})
}

func Param(l Interface, key string, value interface{}) {
	l.Printf("%s = %s\n", key, value)
}

func Scope(l Interface, name string) func() {
	t0 := time.Now()
	l.Printf("start: %s", name)
	return func() {
		l.Printf("finish: %s (time %s)", name, time.Since(t0))
	}
}
