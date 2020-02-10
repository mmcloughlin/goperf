package test

import (
	"flag"
	"testing"
)

var network = flag.Bool("net", false, "allow network access")

func RequiresNetwork(t *testing.T) {
	t.Helper()
	if !*network {
		t.Skip("requires network")
	}
}
