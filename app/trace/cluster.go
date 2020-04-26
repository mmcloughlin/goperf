package trace

import (
	"sort"

	"go.skia.org/infra/perf/go/dataframe"
	"go.skia.org/infra/perf/go/types"
)

func PointsDataFrame(ps []Point) *dataframe.DataFrame {
	df := dataframe.NewEmpty()

	// Construct column headers.
	seen := map[int]bool{}
	for _, p := range ps {
		if seen[p.CommitIndex] {
			continue
		}
		df.Header = append(df.Header, &dataframe.ColumnHeader{
			Offset:    int64(p.CommitIndex),
			Timestamp: p.CommitTime.Unix(),
		})
		seen[p.CommitIndex] = true
	}
	sort.Slice(df.Header, func(i, j int) bool {
		return df.Header[i].Offset < df.Header[j].Offset
	})

	// Map from commit index to column header.
	col := map[int]int{}
	for i, hdr := range df.Header {
		col[int(hdr.Offset)] = i
	}
	n := len(col)

	// Populate trace set.
	for _, p := range ps {
		k := p.key()

		if _, ok := df.TraceSet[k]; !ok {
			df.TraceSet[k] = types.NewTrace(n)
		}

		df.TraceSet[k][col[p.CommitIndex]] = float32(p.Value)
	}

	return df
}
