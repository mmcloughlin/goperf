package results

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/internal/test"
	"github.com/mmcloughlin/cb/pkg/fs"
)

func TestLoaderTestdata(t *testing.T) {
	test.RequiresNetwork(t)

	loader, err := NewLoader(WithFilesystem(fs.NewLocal(".")))
	if err != nil {
		t.Fatal(err)
	}

	filenames, err := filepath.Glob("testdata/*.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range filenames {
		filename := filename // scopelint
		t.Run(filepath.Base(filename), func(t *testing.T) {
			results, err := loader.Load(context.Background(), filename)
			if err != nil {
				t.Fatal(err)
			}

			// Confirm all metadata and environment are the same.
			for _, result := range results {
				if diff := cmp.Diff(results[0].Environment, result.Environment); diff != "" {
					t.Fatalf("environment differs\n%s", diff)
				}
				if diff := cmp.Diff(results[0].Metadata, result.Metadata); diff != "" {
					t.Fatalf("metadata differs\n%s", diff)
				}
			}
		})
	}
}

func TestRel(t *testing.T) {
	cases := []struct {
		Mod, Pkg string
		Expect   string
	}{
		{"golang.org/x/crypto", "golang.org/x/crypto", ""},
		{"golang.org/x/crypto", "golang.org/x/crypto/ssh/terminal", "ssh/terminal"},
	}
	for _, c := range cases {
		r, err := rel(c.Mod, c.Pkg)
		if err != nil {
			t.Fatal(err)
		}
		if r != c.Expect {
			t.Fatalf("got %q; expect %q", r, c.Expect)
		}
	}
}

func TestRelErrors(t *testing.T) {
	cases := []struct {
		Mod, Pkg    string
		ExpectError string
	}{
		{"golang.org/x/sys", "golang.org/x/crypto", `package "golang.org/x/crypto" does not belong to module "golang.org/x/sys"`},
		{"golang.org/x/sys", "golang.org/x/sys\\unix", `expect path separator`},
	}
	for _, c := range cases {
		_, err := rel(c.Mod, c.Pkg)
		if err == nil {
			t.FailNow()
		}
		if err.Error() != c.ExpectError {
			t.Fatalf("got %q; expect %q", err.Error(), c.ExpectError)
		}
	}
}
