// Package trace manipulates benchmark result timeseries.
package trace

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// ID identifies a trace.
type ID struct {
	BenchmarkUUID   uuid.UUID
	EnvironmentUUID uuid.UUID
}

func (i ID) String() string {
	return fmt.Sprintf("%s/%s", i.BenchmarkUUID, i.EnvironmentUUID)
}

// IndexedValue is a measured value at some commit index.
type IndexedValue struct {
	CommitIndex int     `json:"i"`
	Value       float64 `json:"v"`
}

// Series is a series of (commit index, value) pairs in sorted order.
type Series []IndexedValue

func (s Series) Values() []float64 {
	values := make([]float64, len(s))
	for i := range s {
		values[i] = s[i].Value
	}
	return values
}

// Point represents a point in a collection of benchmark timeseries.
type Point struct {
	ID
	IndexedValue
}

// Trace is a timeseries.
type Trace struct {
	ID
	Series Series
}

// Traces gathers points into distinct traces. Values for the same trace ID and
// commit index are averaged.
func Traces(ps []Point) map[ID]*Trace {
	// Gather by (ID, index).
	type key struct {
		ID
		CommitIndex int
	}
	type value struct {
		sum float64
		n   int
	}
	agg := map[key]value{}
	for _, p := range ps {
		k := key{ID: p.ID, CommitIndex: p.CommitIndex}
		v := agg[k]
		v.sum += p.Value
		v.n++
		agg[k] = v
	}

	// Rebuild traces.
	traces := map[ID]*Trace{}
	for k, v := range agg {
		if _, ok := traces[k.ID]; !ok {
			traces[k.ID] = &Trace{ID: k.ID}
		}
		t := traces[k.ID]
		t.Series = append(t.Series, IndexedValue{
			CommitIndex: k.CommitIndex,
			Value:       v.sum / float64(v.n),
		})
	}

	// Sort series by commit index.
	for _, trace := range traces {
		trace := trace // scopelint
		sort.Slice(trace.Series, func(i, j int) bool {
			return trace.Series[i].CommitIndex < trace.Series[j].CommitIndex
		})
	}

	return traces
}
