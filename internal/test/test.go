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

// TempDir creates a temp directory.
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	})

	return dir
}
