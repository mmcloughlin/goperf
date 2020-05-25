package repo

import (
	"context"
	"math/rand"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/goperf/internal/test"
)

var repos = []struct {
	Name string
	Repo Repository
}{
	{
		Name: "gittiles",
		Repo: NewGitilesGo(http.DefaultClient),
	},
	{
		Name: "github",
		Repo: NewGithubGo(http.DefaultClient),
	},
}

func TestRepositoryImplementations(t *testing.T) {
	test.RequiresNetwork(t)

	for _, r := range repos {
		r := r // scopelint
		t.Run(r.Name, func(t *testing.T) {
			// Recent commits.
			commits, err := r.Repo.Log(context.Background(), "HEAD")
			if err != nil {
				t.Fatal(err)
			}

			for _, c := range commits {
				t.Log(c.SHA, c.CommitTime)
			}

			// Lookup a commit by reference.
			commit, err := r.Repo.Revision(context.Background(), "go1.13.3")
			if err != nil {
				t.Fatal(err)
			}

			t.Log(commit.SHA, commit.Author.Name)
		})
	}
}

func TestRepositoryImplementationsSameRevisions(t *testing.T) {
	test.RequiresNetwork(t)

	ctx := context.Background()
	commits, err := repos[0].Repo.Log(ctx, "go1.14beta1")
	if err != nil {
		t.Fatal(err)
	}

	rand.Shuffle(len(commits), func(i, j int) {
		commits[i], commits[j] = commits[j], commits[i]
	})

	if len(commits) > 5 {
		commits = commits[:5]
	}

	for _, commit := range commits {
		t.Logf("commit = %s", commit.SHA)
		expect, err := repos[0].Repo.Revision(ctx, commit.SHA)
		if err != nil {
			t.Fatal(err)
		}

		for _, r := range repos[1:] {
			got, err := r.Repo.Revision(ctx, commit.SHA)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expect, got); diff != "" {
				t.Fatalf("mismatch on commit %s\n%s", commit.SHA, diff)
			}
		}
	}
}
