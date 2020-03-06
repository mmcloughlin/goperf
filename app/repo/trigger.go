package repo

import "time"

type stringvalue struct {
	V string `json:"stringValue"`
}

type timestampvalue struct {
	V time.Time `json:"timestampValue"`
}

type stringarrayvalue struct {
	Array struct {
		Values []stringvalue `json:"values"`
	} `json:"arrayValue"`
}

func (a *stringarrayvalue) Strings() []string {
	var s []string
	for _, sv := range a.Array.Values {
		s = append(s, sv.V)
	}
	return s
}

// CommitFields is a type for receiving commit objects as Cloud Function
// Firestore triggers. Use the Commit() method to convert into a Commit object.
type CommitFields struct {
	triggervalue
}

type triggervalue struct {
	SHA            stringvalue      `json:"sha"`
	Tree           stringvalue      `json:"tree"`
	Parents        stringarrayvalue `json:"parents"`
	AuthorName     stringvalue      `json:"author_name"`
	AuthorEmail    stringvalue      `json:"author_email"`
	AuthorTime     timestampvalue   `json:"author_time"`
	CommitterName  stringvalue      `json:"committer_name"`
	CommitterEmail stringvalue      `json:"committer_email"`
	CommitTime     timestampvalue   `json:"commit_time"`
	Message        stringvalue      `json:"message"`
}

// Commit converts fields to a Commit object.
func (f *CommitFields) Commit() *Commit {
	return &Commit{
		SHA:     f.SHA.V,
		Tree:    f.Tree.V,
		Parents: f.Parents.Strings(),
		Author: Person{
			Name:  f.AuthorName.V,
			Email: f.AuthorEmail.V,
		},
		AuthorTime: f.AuthorTime.V,
		Committer: Person{
			Name:  f.CommitterName.V,
			Email: f.CommitterEmail.V,
		},
		CommitTime: f.CommitTime.V,
		Message:    f.Message.V,
	}
}
