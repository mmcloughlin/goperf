// Package repo provides access to git repository metadata.
package repo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/google/go-github/v29/github"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/gitiles"
)

// GoURL is the URL for the canonical Go git repository.
const GoURL = "https://go.googlesource.com/go"

// Revisions provides query access to repository revisions.
type Revisions interface {
	Revision(ctx context.Context, ref string) (*entity.Commit, error)
}

type revisionscache struct {
	cache *lru.Cache
	r     Revisions
}

// NewRevisionsCache provides an in-memory cache in front of another Revisions fetcher.
func NewRevisionsCache(r Revisions, maxentries int) Revisions {
	return &revisionscache{
		cache: lru.New(maxentries),
		r:     r,
	}
}

func (c *revisionscache) Revision(ctx context.Context, ref string) (*entity.Commit, error) {
	if commit, ok := c.cache.Get(ref); ok {
		return commit.(*entity.Commit), nil
	}

	commit, err := c.r.Revision(ctx, ref)
	if err != nil {
		return nil, err
	}

	c.cache.Add(ref, commit)
	return commit, nil
}

type dbrevisions struct {
	d *db.DB
}

// NewRevisionsFromDatabase looks up revision information from a database. Only
// supports full SHA git refs.
func NewRevisionsFromDatabase(d *db.DB) Revisions {
	return dbrevisions{d: d}
}

func (d dbrevisions) Revision(ctx context.Context, ref string) (*entity.Commit, error) {
	if !isgitsha(ref) {
		return nil, fmt.Errorf("cannot fetch revision information for %q: only supports full git sha refs", ref)
	}
	return d.d.FindCommitBySHA(ctx, ref)
}

// Repository provides access to git repository properties.
type Repository interface {
	Revisions
	Log(ctx context.Context, ref string) ([]*entity.Commit, error)
}

type composite []Repository

// NewCompositeRepository builds a Repository backed by one or more Repository
// implementations. Each method is implemented by calling each sub-Repository in
// turn and returning the first successful result. Panics if no Repositories are
// provided.
func NewCompositeRepository(rs ...Repository) Repository {
	if len(rs) == 0 {
		panic("no repositories provided")
	}
	return composite(rs)
}

func (c composite) Log(ctx context.Context, ref string) (commits []*entity.Commit, err error) {
	for _, r := range c {
		commits, err = r.Log(ctx, ref)
		if err == nil {
			return
		}
	}
	return
}

func (c composite) Revision(ctx context.Context, ref string) (commit *entity.Commit, err error) {
	for _, r := range c {
		commit, err = r.Revision(ctx, ref)
		if err == nil {
			return
		}
	}
	return
}

// Go builds a Repository implementation for the go git repository. API calls
// are made with the provided HTTP client.
func Go(c *http.Client) Repository {
	canonical := NewGitilesGo(c)
	fallback := NewGithubGo(c)
	return NewCompositeRepository(canonical, fallback)
}

type gitilesrepo struct {
	client *gitiles.Client
	repo   string
}

func NewGitiles(c *gitiles.Client, repo string) Repository {
	return &gitilesrepo{
		client: c,
		repo:   repo,
	}
}

// NewGitilesGo builds a Repository implementation for the canonical Go git
// repository at https://go.googlesource.com/go/.
func NewGitilesGo(c *http.Client) Repository {
	gitilesclient := gitiles.NewClient(c, "https://go.googlesource.com")
	return NewGitiles(gitilesclient, "go")
}

func (g *gitilesrepo) Log(ctx context.Context, ref string) ([]*entity.Commit, error) {
	// Fetch repository log.
	res, err := g.client.Log(ctx, g.repo, ref)
	if err != nil {
		return nil, fmt.Errorf("fetching gitiles log: %w", err)
	}

	// Map commits to model type.
	var commits []*entity.Commit
	for _, c := range res.Log {
		commit, err := mapgitilescommit(c)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func (g *gitilesrepo) Revision(ctx context.Context, ref string) (*entity.Commit, error) {
	// Make revision API call.
	res, err := g.client.Revision(ctx, g.repo, ref)
	if err != nil {
		return nil, fmt.Errorf("fetching gitiles revision: %w", err)
	}

	// Map commit to model type.
	c, err := mapgitilescommit(res.Commit)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func mapgitilescommit(c gitiles.Commit) (*entity.Commit, error) {
	// Parse times.
	const timeformat = "Mon Jan _2 15:04:05 2006 -0700"
	authortime, err := time.Parse(timeformat, c.Author.Time)
	if err != nil {
		return nil, fmt.Errorf("author time: %w", err)
	}
	committime, err := time.Parse(timeformat, c.Committer.Time)
	if err != nil {
		return nil, fmt.Errorf("commit time: %w", err)
	}

	// It appears that github does not return trailing whitespace in commit
	// messages. Trim here to match.
	message := strings.TrimSpace(c.Message)

	// Convert into model type.
	return &entity.Commit{
		SHA:     c.SHA,
		Tree:    c.Tree,
		Parents: c.Parents,
		Author: entity.Person{
			Name:  c.Author.Name,
			Email: c.Author.Email,
		},
		AuthorTime: authortime,
		Committer: entity.Person{
			Name:  c.Committer.Name,
			Email: c.Committer.Email,
		},
		CommitTime: committime,
		Message:    message,
	}, nil
}

type githubrepo struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGithub(c *github.Client, owner, repo string) Repository {
	return &githubrepo{
		client: c,
		owner:  owner,
		repo:   repo,
	}
}

// NewGithubGo builds a Repository implementation for the github mirror of the
// Go git repository at https://github.com/golang/go.
func NewGithubGo(c *http.Client) Repository {
	githubclient := github.NewClient(c)
	return NewGithub(githubclient, "golang", "go")
}

func (g *githubrepo) Log(ctx context.Context, ref string) ([]*entity.Commit, error) {
	// List commits.
	res, _, err := g.client.Repositories.ListCommits(ctx, g.owner, g.repo, &github.CommitsListOptions{
		SHA: ref,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching github commits: %w", err)
	}

	// Map commits to model type.
	var commits []*entity.Commit
	for _, c := range res {
		commits = append(commits, mapgithubcommit(c))
	}

	return commits, nil
}

func (g *githubrepo) Revision(ctx context.Context, ref string) (*entity.Commit, error) {
	// Make commit API call.
	res, _, err := g.client.Repositories.GetCommit(ctx, g.owner, g.repo, ref)
	if err != nil {
		return nil, fmt.Errorf("fetching github revision: %w", err)
	}

	// Map commit to model type.
	return mapgithubcommit(res), nil
}

func mapgithubcommit(c *github.RepositoryCommit) *entity.Commit {
	var parents []string
	for _, parent := range c.Parents {
		parents = append(parents, parent.GetSHA())
	}

	return &entity.Commit{
		SHA:     c.GetSHA(),
		Tree:    c.GetCommit().GetTree().GetSHA(),
		Parents: parents,
		Author: entity.Person{
			Name:  c.GetCommit().GetAuthor().GetName(),
			Email: c.GetCommit().GetAuthor().GetEmail(),
		},
		AuthorTime: c.GetCommit().GetAuthor().GetDate(),
		Committer: entity.Person{
			Name:  c.GetCommit().GetCommitter().GetName(),
			Email: c.GetCommit().GetCommitter().GetEmail(),
		},
		CommitTime: c.GetCommit().GetCommitter().GetDate(),
		Message:    c.GetCommit().GetMessage(),
	}
}
