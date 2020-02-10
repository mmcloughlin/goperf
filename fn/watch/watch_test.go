package watch

import (
	"context"
	"flag"
	"testing"
)

var network = flag.Bool("net", false, "allow network access")

func TestRecentCommits(t *testing.T) {
	if !*network {
		t.Skip("requires network")
	}

	commits, err := RecentCommits(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range commits {
		t.Log(c.Commit, c.Committer.Time)
	}
}
