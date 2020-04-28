package change

import (
	"fmt"
	"math"

	analysis "golang.org/x/perf/analysis/app"

	"github.com/mmcloughlin/cb/app/trace"
)

type Detector interface {
	Name() string
	Detect(series trace.Series) []Change
}

type KZA struct {
	M, K             int
	PercentThreshold float64
}

func (k *KZA) Name() string {
	return fmt.Sprintf("kza(%v,%v,%v%%)", k.M, k.K, k.PercentThreshold)
}

func (k *KZA) Detect(series trace.Series) []Change {
	var changes []Change

	f := analysis.AdaptiveKolmogorovZurbenko(series.Values(), k.M, k.K)

	for i := 1; i < len(f); i++ {
		pre, post := f[i-1], f[i]
		percent := 100 * math.Abs((post-pre)/pre)
		if percent > k.PercentThreshold {
			changes = append(changes, Change{
				CommitIndex: series[i].CommitIndex,
				EffectSize:  percent,
			})
		}
	}

	return changes
}

type Cohen struct {
	WindowSize    int     // window to consider either side
	MinEffectSize float64 // Cohen's d threshold
}

func (c *Cohen) Name() string {
	return fmt.Sprintf("cohen(%v,%v)", c.WindowSize, c.MinEffectSize)
}

func (c *Cohen) Detect(series trace.Series) []Change {
	var changes []Change

	// Initialize window statistics.
	n := len(series)
	w := newwindows()
	for _, v := range series {
		w.push(v.Value)
	}

	// Consider each point with a full window either side.
	for i := c.WindowSize; i+c.WindowSize <= n; i++ {
		// Pre and post statistics.
		pre := w.stats(i-c.WindowSize, i)
		post := w.stats(i, i+c.WindowSize)

		// Cohen's d effect size.
		effect := cohen(pre, post)
		if math.Abs(effect) > c.MinEffectSize {
			changes = append(changes, Change{
				CommitIndex: series[i].CommitIndex,
				EffectSize:  effect,
			})
		}
	}

	return changes
}

type Hybrid struct {
	WindowSize    int     // window to consider either side
	MinEffectSize float64 // Cohen's d threshold

	M, K             int     // KZA k parameter
	PercentThreshold float64 // threshold for KZA pass
	Context          int     // number of points to consider either side
}

func (h *Hybrid) Name() string { return "hybrid" }

func (h *Hybrid) Detect(series trace.Series) []Change {
	var changes []Change

	values := series.Values()

	w := newwindows()
	w.push(values...)

	// Pre-process with KZA.
	f := analysis.AdaptiveKolmogorovZurbenko(values, h.M, h.K)

	for i := 1; i < len(f); i++ {
		percent := 100 * math.Abs((f[i]-f[i-1])/f[i-1])
		if percent < h.PercentThreshold {
			continue
		}

		// Find largest effect size in a small window around this candidate.
		chg := Change{}
		for j := i - h.Context; j <= i+h.Context; j++ {
			if j < h.WindowSize || j+h.WindowSize >= len(values) {
				continue
			}
			pre := w.stats(j-h.WindowSize, j)
			post := w.stats(j, j+h.WindowSize)
			effect := cohen(pre, post)
			if math.Abs(effect) > math.Abs(chg.EffectSize) {
				chg.CommitIndex = series[j].CommitIndex
				chg.EffectSize = effect
			}
		}

		if math.Abs(chg.EffectSize) > h.MinEffectSize {
			changes = append(changes, chg)
		}
	}

	return changes
}
