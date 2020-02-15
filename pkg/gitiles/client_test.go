package gitiles

import (
	"context"
	"net/http"
	"testing"

	"github.com/mmcloughlin/cb/test"
)

func TestClientLog(t *testing.T) {
	test.RequiresNetwork(t)

	c := NewClient(http.DefaultClient, "https://go.googlesource.com")
	r, err := c.Log(context.Background(), "go")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range r.Log {
		t.Log(c.Commit, c.Committer.Time)
	}
}
