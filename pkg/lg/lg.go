package lg

import (
	"log"
	"os"
	"testing"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

func Default() Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

type noop struct{}

func (noop) Printf(format string, v ...interface{}) {}

func Noop() Logger { return noop{} }

type test struct {
	t *testing.T
}

func Test(t *testing.T) Logger { return test{t} }

func (t test) Printf(format string, v ...interface{}) {
	t.t.Logf(format, v...)
}

func Param(l Logger, key string, value interface{}) {
	l.Printf("%s = %v\n", key, value)
}

func Scope(l Logger, name string) func() {
	t0 := time.Now()
	l.Printf("start: %s", name)
	return func() {
		l.Printf("finish: %s (time %s)", name, time.Since(t0))
	}
}

func Error(l Logger, name string, err error) {
	if err != nil {
		l.Printf("%s: %v", name, err)
	}
}
