package change

import (
	"flag"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/goperf/app/change/changetest"
	"github.com/mmcloughlin/goperf/app/trace"
)

var plot = flag.Bool("plot", false, "generate plots for test cases")

func TestDetectTestData(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	detector := DefaultDetector

	for _, filename := range filenames {
		filename := filename // scopelint
		name := filepath.Base(filename)
		t.Run(name, func(t *testing.T) {
			// Read test case.
			tc, err := changetest.ReadCaseFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			// Debug plot.
			if *plot {
				plotname := strings.TrimSuffix(filename, ".json") + ".png"
				if err := changetest.PlotSeries(plotname, name, tc.Series); err != nil {
					t.Fatal(err)
				}
			}

			// Detect changes.
			changes := detector.Detect(tc.Series)
			LogChanges(t, changes)

			// Extract change points.
			points := []int{}
			for _, c := range changes {
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
	LogChanges(t, changes)

	// Expect change at 100.
	AssertOneChangeAt(t, changes, 100)
}

func TestDetectWindowClipped(t *testing.T) {
	detector := DefaultDetector

	// Test case with *massive* step change, but a window on one side that's
	// "clipped", meaning smaller than the window of the detector. It's possible
	// in this case that the change will still be detected, but not at the right
	// position.

	w := detector.WindowSize

	// Generate an absurdly obvious step change, but with not enough of a window
	// on one side.
	var series trace.Series
	series = AppendRandNormSeries(series, 17, 1, 100)
	series = AppendRandNormSeries(series, 100, 1, w-3)

	// Detect changes.
	changes := detector.Detect(series)
	LogChanges(t, changes)

	AssertOneChangeAt(t, changes, 100)
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

func LogChanges(t *testing.T, changes []Change) {
	for _, c := range changes {
		t.Logf("%d: %v", c.CommitIndex, c.EffectSize)
	}
}

func AssertOneChangeAt(t *testing.T, changes []Change, expect int) {
	if len(changes) != 1 {
		t.Fatalf("expect 1 change; got %d", len(changes))
	}

	change := changes[0]

	if change.CommitIndex != expect {
		t.Fatalf("expected change at index %d; got %d", expect, change.CommitIndex)
	}
}
