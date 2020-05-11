package change

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/app/change/changetest"
	"github.com/mmcloughlin/cb/app/trace"
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

func TestDetectGenerated(t *testing.T) {
	// Sanity check on artificially generated step function.
	var series trace.Series
	series = AppendRandNormSeries(series, 17, 1, 100)
	series = AppendRandNormSeries(series, 42, 1, 100)

	// Detect changes.
	changes := DefaultDetector.Detect(series)

	if len(changes) != 1 {
		t.Fatalf("expect 1 change; got %d", len(changes))
	}

	change := changes[0]
	t.Logf("change = %#v", change)

	if change.CommitIndex != 100 {
		t.Fatalf("expected change at index 100; got %d", change.CommitIndex)
	}
}

// RandNorm samples from normal distribution with mean m and standard deviation
// s.
func RandNorm(m, s float64) float64 {
	return m + s*rand.NormFloat64()
}

func AppendRandNormSeries(series trace.Series, m, s float64, n int) trace.Series {
	idx := 0
	if len(series) != 0 {
		idx = series[len(series)-1].CommitIndex + 1
	}

	for i := 0; i < n; i++ {
		series = append(series, trace.IndexedValue{
			CommitIndex: idx,
			Value:       RandNorm(m, s),
		})
		idx++
	}

	return series
}
