package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcloughlin/cb/gitiles"
)

// Log provides access to a git repository commit log.
type Log interface {
	RecentCommits(ctx context.Context) ([]Commit, error)
}

type gitileslog struct {
	client *gitiles.Client
	repo   string
}

func NewGitilesLog(c *gitiles.Client, repo string) Log {
	return &gitileslog{
		client: c,
		repo:   repo,
	}
}

func (g *gitileslog) RecentCommits(ctx context.Context) ([]Commit, error) {
	// Fetch repository log.
	res, err := g.client.Log(ctx, g.repo)
	if err != nil {
		return nil, fmt.Errorf("fetching gitiles log: %w", err)
	}

	// Map commits to model type.
	var commits []Commit
	for _, c := range res.Log {
		commit, err := mapgitilescommit(c)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func mapgitilescommit(c gitiles.Commit) (Commit, error) {
	// Parse times.
	const timeformat = "Mon Jan _2 15:04:05 2006 -0700"
	authortime, err := time.Parse(timeformat, c.Author.Time)
	if err != nil {
		return Commit{}, fmt.Errorf("author time: %w", err)
	}
	committime, err := time.Parse(timeformat, c.Committer.Time)
	if err != nil {
		return Commit{}, fmt.Errorf("commit time: %w", err)
	}

	// Convert into model type.
	return Commit{
		SHA:     c.Commit,
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
