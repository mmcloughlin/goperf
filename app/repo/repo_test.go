package repo

import (
	"context"
	"net/http"
	"testing"

	"github.com/mmcloughlin/cb/internal/test"
	"github.com/mmcloughlin/cb/pkg/gitiles"
)

func TestGitiles(t *testing.T) {
	test.RequiresNetwork(t)

	c := gitiles.NewClient(http.DefaultClient, "https://go.googlesource.com")
	l := NewGitiles(c, "go")

	// Recent commits.
	commits, err := l.RecentCommits(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range commits {
		t.Log(c.SHA, c.CommitTime)
	}

	// Lookup a commit by reference.
	commit, err := l.Revision(context.Background(), "go1.13.3")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(commit.SHA, commit.Author.Name)
}
