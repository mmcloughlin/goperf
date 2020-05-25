package gitiles

import (
	"context"
	"net/http"
	"testing"

	"github.com/mmcloughlin/goperf/internal/test"
)

func TestClientLog(t *testing.T) {
	test.RequiresNetwork(t)

	c := NewClient(http.DefaultClient, "https://go.googlesource.com")
	r, err := c.Log(context.Background(), "go", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range r.Log {
		t.Log(c.SHA, c.Committer.Time)
	}
}

func TestClientRevision(t *testing.T) {
	test.RequiresNetwork(t)

	c := NewClient(http.DefaultClient, "https://go.googlesource.com")
	r, err := c.Revision(context.Background(), "go", "go1.13.7")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)

	if r.SHA != "7d2473dc81c659fba3f3b83bc6e93ca5fe37a898" {
		t.Fatal("unexpected commit sha")
	}
}
