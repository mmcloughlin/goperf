package repo

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/mmcloughlin/cb/app/entity"
)

// Git is a cloned local git repository.
type Git struct {
	r   *git.Repository
	dir string
}

// Clone a repository to a temporary directory.
func Clone(ctx context.Context, url string) (*Git, error) {
	d, err := ioutil.TempDir("", "clone")
	if err != nil {
		return nil, err
	}

	r, err := git.PlainCloneContext(ctx, d, true, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		_ = os.RemoveAll(d)
		return nil, err
	}

	return &Git{
		r:   r,
		dir: d,
	}, nil
}

// Close deletes the cloned repository.
func (g *Git) Close() error {
	return os.RemoveAll(g.dir)
}

// Log is equivalent to "git log".
func (g *Git) Log() (CommitIterator, error) {
	ref, err := g.r.Head()
	if err != nil {
		return nil, err
	}

	l, err := g.r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}

	return commitIterator{l}, nil
}

// CommitIterator provides iterative access to a sequence of commits.
type CommitIterator interface {
	// Next returns the next commit, or io.EOF when done.
	Next() (*entity.Commit, error)
	Close()
}

type commitIterator struct {
	object.CommitIter
}

func (i commitIterator) Next() (*entity.Commit, error) {
	c, err := i.CommitIter.Next()
	if err != nil {
		return nil, err
	}
	return mapGitCommit(c), nil
}

func mapGitCommit(c *object.Commit) *entity.Commit {
	parents := make([]string, len(c.ParentHashes))
	for i, parent := range c.ParentHashes {
		parents[i] = parent.String()
	}
	return &entity.Commit{
		SHA:        c.Hash.String(),
		Tree:       c.TreeHash.String(),
		Parents:    parents,
		Author:     mapGitSignatureToPerson(c.Author),
		AuthorTime: c.Author.When,
		Committer:  mapGitSignatureToPerson(c.Committer),
		CommitTime: c.Committer.When,
		Message:    c.Message,
	}
}

func mapGitSignatureToPerson(s object.Signature) entity.Person {
	return entity.Person{
		Name:  s.Name,
		Email: s.Email,
	}
}

type firstParent struct {
	CommitIterator
	prev *entity.Commit
}

// FirstParent filters the given commit sequence by only following the first parent at merge commits.
func FirstParent(commits CommitIterator) CommitIterator {
	return &firstParent{
		CommitIterator: commits,
	}
}

func (f *firstParent) Next() (*entity.Commit, error) {
	// First iteration.
	if f.prev == nil {
		c, err := f.CommitIterator.Next()
		if err != nil {
			return nil, err
		}
		f.prev = c
		return c, nil
	}

	// Look for commit matching first parent of the previous one.
	if len(f.prev.Parents) == 0 {
		return nil, io.EOF
	}

	for {
		c, err := f.CommitIterator.Next()
		if err != nil {
			return nil, err
		}

		if c.SHA == f.prev.Parents[0] {
			f.prev = c
			return c, nil
		}
	}
}
