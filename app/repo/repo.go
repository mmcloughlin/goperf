package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/google/go-github/github"

	"github.com/mmcloughlin/cb/pkg/gitiles"
)

// Revisions provides query access to repository revisions.
type Revisions interface {
	Revision(ctx context.Context, ref string) (*Commit, error)
}

type revisionscache struct {
	cache *lru.Cache
	r     Revisions
}

// NewRevisionsCache provides an in-memory cache in front of another Revisions fetcher.
func NewRevisionsCache(r Revisions, maxentries int) Revisions {
	return revisionscache{
		cache: lru.New(maxentries),
		r:     r,
	}
}

func (c revisionscache) Revision(ctx context.Context, ref string) (*Commit, error) {
	if commit, ok := c.cache.Get(ref); ok {
		return commit.(*Commit), nil
	}

	commit, err := c.r.Revision(ctx, ref)
	if err != nil {
		return nil, err
	}

	c.cache.Add(ref, commit)
	return commit, nil
}

// Repository provides access to git repository properties.
type Repository interface {
	Revisions
	RecentCommits(ctx context.Context) ([]*Commit, error)
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

func (g *gitilesrepo) RecentCommits(ctx context.Context) ([]*Commit, error) {
	// Fetch repository log.
	res, err := g.client.Log(ctx, g.repo)
	if err != nil {
		return nil, fmt.Errorf("fetching gitiles log: %w", err)
	}

	// Map commits to model type.
	var commits []*Commit
	for _, c := range res.Log {
		commit, err := mapgitilescommit(c)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func (g *gitilesrepo) Revision(ctx context.Context, ref string) (*Commit, error) {
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

func mapgitilescommit(c gitiles.Commit) (*Commit, error) {
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

	// Convert into model type.
	return &Commit{
		SHA:     c.SHA,
		Tree:    c.Tree,
		Parents: c.Parents,
		Author: Person{
			Name:  c.Author.Name,
			Email: c.Author.Email,
		},
		AuthorTime: authortime,
		Committer: Person{
			Name:  c.Committer.Name,
			Email: c.Committer.Email,
		},
		CommitTime: committime,
		Message:    c.Message,
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

func (g *githubrepo) RecentCommits(ctx context.Context) ([]*Commit, error) {
	// List commits.
	res, _, err := g.client.Repositories.ListCommits(ctx, g.owner, g.repo, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching github commits: %w", err)
	}

	// Map commits to model type.
	var commits []*Commit
	for _, c := range res {
		commits = append(commits, mapgithubcommit(c))
	}

	return commits, nil
}

func (g *githubrepo) Revision(ctx context.Context, ref string) (*Commit, error) {
	// Make commit API call.
	res, _, err := g.client.Repositories.GetCommit(ctx, g.owner, g.repo, ref)
	if err != nil {
		return nil, fmt.Errorf("fetching github revision: %w", err)
	}

	// Map commit to model type.
	return mapgithubcommit(res), nil
}

func mapgithubcommit(c *github.RepositoryCommit) *Commit {
	var parents []string
	for _, parent := range c.GetCommit().Parents {
		parents = append(parents, parent.GetSHA())
	}

	return &Commit{
		SHA:     c.GetSHA(),
		Tree:    c.GetCommit().GetTree().GetSHA(),
		Parents: parents,
		Author: Person{
			Name:  c.GetCommit().GetAuthor().GetName(),
			Email: c.GetCommit().GetAuthor().GetEmail(),
		},
		AuthorTime: c.GetCommit().GetAuthor().GetDate(),
		Committer: Person{
			Name:  c.GetCommit().GetCommitter().GetName(),
			Email: c.GetCommit().GetCommitter().GetEmail(),
		},
		CommitTime: c.GetCommit().GetCommitter().GetDate(),
		Message:    c.GetCommit().GetMessage(),
	}
}
