package httputil

import "testing"

func TestParseNameVersion(t *testing.T) {
	cases := []struct {
		Path    string
		Name    string
		Version string
	}{
		{
			Path:    "/a/b/c/file.0123456789ab.ext",
			Name:    "/a/b/c/file.ext",
			Version: "0123456789ab",
		},
		{
			Path:    "/a/b/c/file.old.ext",
			Name:    "/a/b/c/file.old.ext",
			Version: "",
		},
		{
			Path:    "/a/b/c/file.ext",
			Name:    "/a/b/c/file.ext",
			Version: "",
		},
		{
			Path:    "/a/b/c/file",
			Name:    "/a/b/c/file",
			Version: "",
		},
		{
			Path:    "",
			Name:    "",
			Version: "",
		},
	}
	for _, c := range cases {
		name, version := parseNameVersion(c.Path)
		if name != c.Name || version != c.Version {
			t.Fatalf("parseNameVersion(%v) = %v, %v; expect %v, %v", c.Path, name, version, c.Name, c.Version)
		}
	}
}

func TestVersionedPathRoundtrip(t *testing.T) {
	path := "/a/b/c/file.0123456789ab.ext"
	got := versionedPath(parseNameVersion(path))
	if got != path {
		t.Fatalf("roundtrip failed: got %q expect %q", got, path)
	}
}
