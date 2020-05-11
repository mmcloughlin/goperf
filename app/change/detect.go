package change

import (
	"math"

	analysis "golang.org/x/perf/analysis/app"

	"github.com/mmcloughlin/cb/app/trace"
)

// Detector is a change detector.
//
// Uses a hybrid approach. A first pass Adaptive Kolmogorov-Zurbenko (KZA)
// filter is applied to identify structural breaks in the timeseries. This is
// effective at identifiying regions of interest, but imprecise at pinpointing
// the exact change point. For each candidates from the KZA pass, we inspect a
// few points around the candidate and compare distributions of a windows either
// size. The point with the largest effect size (Cohen's d) is taken as the
// change point.
type Detector struct {
	// Adaptive Kolmogorov-Zurbenko pass.
	M, K             int     // KZA parameters
	PercentThreshold float64 // threshold for KZA pass
	Context          int     // number of points to consider either side

	// Distribution comparison.
	WindowSize    int     // window to consider either side
	MinEffectSize float64 // Cohen's d threshold
}

// DefaultDetector has sensible default parameter choices.
var DefaultDetector = &Detector{
	WindowSize:    20,
	MinEffectSize: 3,

	M:                15,
	K:                3,
	PercentThreshold: 4,
	Context:          2,
}

// Detect changes in series.
func (d *Detector) Detect(series trace.Series) []Change {
	var changes []Change

	values := series.Values()

	w := newwindows()
	w.push(values...)

	// Pre-process with KZA.
	f := analysis.AdaptiveKolmogorovZurbenko(values, d.M, d.K)

	hasChange := map[int]bool{}
	for i := 1; i < len(f); i++ {
		percent := 100 * math.Abs((f[i]-f[i-1])/f[i-1])
		if percent < d.PercentThreshold {
			continue
		}

		// Find largest effect size in a small window around this candidate.
		chg := Change{}
		for j := i - d.Context; j <= i+d.Context; j++ {
			pre := w.stats(max(j-d.WindowSize, 0), j)
			post := w.stats(j, min(j+d.WindowSize, len(values)))
			effect := cohen(post, pre)
			if math.Abs(effect) > math.Abs(chg.EffectSize) {
				chg.CommitIndex = series[j].CommitIndex
				chg.EffectSize = effect
				chg.Pre = pre
				chg.Post = post
			}
		}

		if math.Abs(chg.EffectSize) > d.MinEffectSize && !hasChange[chg.CommitIndex] {
			changes = append(changes, chg)
			hasChange[chg.CommitIndex] = true
		}
	}

	return changes
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
