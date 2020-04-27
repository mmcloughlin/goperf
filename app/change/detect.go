package change

import (
	"math"

	"github.com/mmcloughlin/cb/app/trace"
)

type Detector struct {
	WindowSize    int     // window to consider either side
	MinEffectSize float64 // Cohen's d threshold
}

func (d *Detector) Detect(series trace.Series) []Change {
	var changes []Change

	// Initialize window statistics.
	n := len(series)
	w := newwindows()
	for _, v := range series {
		w.push(v.Value)
	}

	// Consider each point with a full window either side.
	for i := d.WindowSize; i+d.WindowSize <= n; i++ {
		// Pre and post statistics.
		pre := w.stats(i-d.WindowSize, i)
		post := w.stats(i, i+d.WindowSize)

		// Cohen's d effect size.
		effect := cohen(pre, post)
		if math.Abs(effect) > d.MinEffectSize {
			changes = append(changes, Change{
				CommitIndex: series[i].CommitIndex,
				EffectSize:  effect,
				Pre:         pre,
				Post:        post,
			})
		}
	}

	return changes
}

// cohen computes Cohen's d effect size between two means.
func cohen(s1, s2 Stats) float64 {
	return (s1.Mean - s2.Mean) / pooledStddev(s1, s2)
}

// pooledVariance computes the pooled variance over two samples.
func pooledVariance(s1, s2 Stats) float64 {
	n1 := float64(s1.N - 1)
	n2 := float64(s2.N - 1)
	return (n1*s1.Variance + n2*s2.Variance) / (n1 + n2)
}

// pooledStddev computes the pooled standard deviation over two samples.
func pooledStddev(s1, s2 Stats) float64 {
	return math.Sqrt(pooledVariance(s1, s2))
}

// windows assists with computing statistics for windows in a sequence.
type windows struct {
	n      int
	cumlx  []float64 // cumlx[i] = sum of x[j] for j < i
	cumlx2 []float64 // cumlx2[i] = sum of x[j]^2 for j < i
}

// newwindows initializes an empty windows sequence.
func newwindows() *windows {
	return &windows{
		n:      0,
		cumlx:  []float64{0},
		cumlx2: []float64{0},
	}
}

// push value at the end of the sequence.
func (w *windows) push(x float64) {
	w.cumlx = append(w.cumlx, w.cumlx[w.n]+x)
	w.cumlx2 = append(w.cumlx2, w.cumlx2[w.n]+x*x)
	w.n++
}

// sum of window x[l:r].
func (w *windows) sum(l, r int) float64 {
	return w.cumlx[r] - w.cumlx[l]
}

// sumsq returns sum of squares in window x[l:r].
func (w *windows) sumsq(l, r int) float64 {
	return w.cumlx2[r] - w.cumlx2[l]
}

// mean of the window x[l:r].
func (w *windows) mean(l, r int) float64 {
	return w.sum(l, r) / float64(r-l)
}

// sampvar returns the sample variance of the window x[l:r].
func (w *windows) sampvar(l, r int) float64 {
	sumsq := w.sumsq(l, r)
	sum := w.sum(l, r)
	n := float64(r - l)
	return (sumsq - sum*sum/n) / (n - 1)
}

// stats for the window x[l:r].
func (w *windows) stats(l, r int) Stats {
	return Stats{
		N:        r - l,
		Mean:     w.mean(l, r),
		Variance: w.sampvar(l, r),
	}
}
