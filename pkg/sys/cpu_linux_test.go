package sys

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

var update = flag.Bool("update", false, "update golden files")

func TestCPUReal(t *testing.T) {
	c, err := CPU.Configuration()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCPUGolden(t *testing.T) {
	filenames, err := filepath.Glob("testdata/cpuinfo_*.input")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		filename := filename
		base := filepath.Base(filename)
		t.Run(base, func(t *testing.T) {
			// Parse.
			c, err := cpucfg(filename, cfg.TagPerfCritical)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.Validate(); err != nil {
				t.Fatal(err)
			}

			// Write out config.
			buf := bytes.NewBuffer(nil)
			if err := cfg.Write(buf, c); err != nil {
				t.Fatal(err)
			}
			got := buf.Bytes()

			// Update golden file if requested.
			golden := strings.ReplaceAll(filename, ".input", ".golden")
			if *update {
				if err := ioutil.WriteFile(golden, got, 0o666); err != nil {
					t.Fatal(err)
				}
			}

			// Read golden file.
			expect, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expect, got); diff != "" {
				t.Fatalf("mismatch\n%s", diff)
			}
		})
	}
}
