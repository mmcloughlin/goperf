package change

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/app/change/changetest"
)

func TestDetectTestData(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	detector := DefaultDetector

	for _, filename := range filenames {
		filename := filename // scopelint
		t.Run(filepath.Base(filename), func(t *testing.T) {
			// Read test case.
			tc, err := changetest.ReadCaseFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			// Detect changes.
			changes := detector.Detect(tc.Series)

			// Extract change points.
			points := []int{}
			for _, c := range changes {
				t.Logf("%d: %v", c.CommitIndex, c.EffectSize)
				points = append(points, c.CommitIndex)
			}

			if diff := cmp.Diff(tc.Expect, points); diff != "" {
				t.Errorf("mismatch\n%s", diff)
			}
		})
	}
}
