package model

import "time"

type Commit struct {
	SHA            string    `firestore:"sha" json:"sha"`
	Tree           string    `firestore:"tree" json:"tree"`
	Parents        []string  `firestore:"parents" json:"parents"`
	AuthorName     string    `firestore:"author_name" json:"author_name"`
	AuthorEmail    string    `firestore:"author_email" json:"author_email"`
	AuthorTime     time.Time `firestore:"author_time" json:"author_time"`
	CommitterName  string    `firestore:"committer_name" json:"committer_name"`
	CommitterEmail string    `firestore:"committer_email" json:"committer_email"`
	CommitTime     time.Time `firestore:"commit_time" json:"commit_time"`
	Message        string    `firestore:"message" json:"message"`
}

func (c *Commit) Type() string { return "commits" }

func (c *Commit) ID() string { return c.SHA }
