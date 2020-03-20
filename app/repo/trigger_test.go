package repo

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/mmcloughlin/cb/app/entity"
)

var triggerpayload = `{
	"author_email": {
		"stringValue": "joel@sing.id.au"
	},
	"author_name": {
		"stringValue": "Joel Sing"
	},
	"author_time": {
		"timestampValue": "2020-03-01T17:26:54Z"
	},
	"commit_time": {
		"timestampValue": "2020-03-05T11:56:33Z"
	},
	"committer_email": {
		"stringValue": "joel@sing.id.au"
	},
	"committer_name": {
		"stringValue": "Joel Sing"
	},
	"message": {
		"stringValue": "cmd/compile: add zero store operations for riscv64"
	},
	"parents": {
		"arrayValue": {
			"values": [
				{
					"stringValue": "585e31df63f6879c03b285711de6f9dcba1f2cb0"
				}
			]
		}
	},
	"sha": {
		"stringValue": "cc6a8bd0d7f782c31e1a35793b4e1253c6716ad5"
	},
	"tree": {
		"stringValue": "efa562f9aafdc87a201acbdd3da66ee0ec6587ee"
	}
}`

func TestCommitFieldsCommit(t *testing.T) {
	// Unmarshal.
	var f CommitFields
	if err := json.Unmarshal([]byte(triggerpayload), &f); err != nil {
		t.Fatal(err)
	}

	got := f.Commit()
	expect := &entity.Commit{
		SHA:     "cc6a8bd0d7f782c31e1a35793b4e1253c6716ad5",
		Tree:    "efa562f9aafdc87a201acbdd3da66ee0ec6587ee",
		Parents: []string{"585e31df63f6879c03b285711de6f9dcba1f2cb0"},
		Author: entity.Person{
			Name:  "Joel Sing",
			Email: "joel@sing.id.au",
		},
		AuthorTime: time.Date(2020, 3, 1, 17, 26, 54, 0, time.UTC),
		Committer: entity.Person{
			Name:  "Joel Sing",
			Email: "joel@sing.id.au",
		},
		CommitTime: time.Date(2020, 3, 5, 11, 56, 33, 0, time.UTC),
		Message:    "cmd/compile: add zero store operations for riscv64",
	}

	if !reflect.DeepEqual(expect, got) {
		t.Logf("got =\n%#v", got)
		t.Logf("expect =\n%#v", expect)
		t.Fail()
	}
}
