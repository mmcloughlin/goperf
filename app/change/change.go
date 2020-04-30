// Package change implements change detection in benchmark timeseries.
package change

type Change struct {
	CommitIndex int
	EffectSize  float64
	Pre         Stats
	Post        Stats
}
