package test

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

var network = flag.Bool("net", false, "allow network access")

func RequiresNetwork(t *testing.T) {
	t.Helper()
	if !*network {
		t.Skip("requires network")
	}
}

// TempDir creates a temp directory. Returns the path to the directory and a
// cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
}
