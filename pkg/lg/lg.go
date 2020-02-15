package lg

import (
	"log"
	"os"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

func Default() Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

func Param(l Logger, key string, value interface{}) {
	l.Printf("%s = %s\n", key, value)
}

func Scope(l Logger, name string) func() {
	t0 := time.Now()
	l.Printf("start: %s", name)
	return func() {
		l.Printf("finish: %s (time %s)", name, time.Since(t0))
	}
}
