// Package mapper maps internal types to models.
package mapper

import (
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/model"
)

func CommitToModel(c *entity.Commit) *model.Commit {
	return &model.Commit{
		SHA:            c.SHA,
		Tree:           c.Tree,
		Parents:        c.Parents,
		AuthorName:     c.Author.Name,
		AuthorEmail:    c.Author.Email,
		AuthorTime:     c.AuthorTime,
		CommitterName:  c.Committer.Name,
		CommitterEmail: c.Committer.Email,
		CommitTime:     c.CommitTime,
		Message:        c.Message,
	}
}

func CommitFromModel(c *model.Commit) *entity.Commit {
	return &entity.Commit{
		SHA:     c.SHA,
		Tree:    c.Tree,
		Parents: c.Parents,
		Author: entity.Person{
			Name:  c.AuthorName,
			Email: c.AuthorEmail,
		},
		AuthorTime: c.AuthorTime,
		Committer: entity.Person{
			Name:  c.CommitterName,
			Email: c.CommitterEmail,
		},
		CommitTime: c.CommitTime,
		Message:    c.Message,
	}
}
