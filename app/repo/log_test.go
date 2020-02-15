package repo

import (
	"context"
	"net/http"
	"testing"

	"github.com/mmcloughlin/cb/pkg/gitiles"
	"github.com/mmcloughlin/cb/pkg/test"
)

func TestGitilesLog(t *testing.T) {
	test.RequiresNetwork(t)

	c := gitiles.NewClient(http.DefaultClient, "https://go.googlesource.com")
	l := NewGitilesLog(c, "go")

	commits, err := l.RecentCommits(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range commits {
		t.Log(c.SHA, c.CommitTime)
	}
}
