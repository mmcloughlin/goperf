package watch

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	// Fetch commits.
	commits, err := RecentCommits(r.Context())
	if err != nil {
		log.Printf("recent commits: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log results.
	for _, c := range commits {
		log.Println(c.Commit, c.Committer.Time)
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}

type Ident struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Time  string `json:"time"`
}

type Commit struct {
	Commit    string   `json:"commit"`
	Tree      string   `json:"tree"`
	Parents   []string `json:"parents"`
	Author    Ident    `json:"author"`
	Committer Ident    `json:"committer"`
	Message   string   `json:"message"`
}

var client = &http.Client{}

func RecentCommits(ctx context.Context) ([]Commit, error) {
	// Build request.
	u := "https://go.googlesource.com/go/+log?format=JSON"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	// Execute the request.
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Parse response body.
	var payload struct {
		Log  []Commit `json:"log"`
		Next string   `json:"next"`
	}
	if err := DecodeJSON(res.Body, &payload); err != nil {
		return nil, err
	}

	return payload.Log, nil
}

func DecodeJSON(rd io.Reader, v interface{}) error {
	r := bufio.NewReader(rd)

	// Expect response to start with "magic" byte sequence.
	magic := []byte(")]}'\n")
	prefix, err := r.Peek(len(magic))
	if err != nil {
		return err
	}
	if !bytes.Equal(magic, prefix) {
		return fmt.Errorf("expected response body to start with magic %q; got %q", magic, prefix)
	}
	r.Discard(len(magic))

	// Decode as JSON.
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()

	if err := d.Decode(v); err != nil {
		return err
	}

	// Should not have trailing data.
	if d.More() {
		return errors.New("unexpected extra data after JSON")
	}

	return nil
}
