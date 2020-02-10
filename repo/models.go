package repo

import "time"

type Person struct {
	Name  string
	Email string
}

type Commit struct {
	SHA        string
	Tree       string
	Parents    []string
	Author     Person
	AuthorTime time.Time
	Committer  Person
	CommitTime time.Time
	Message    string
}
