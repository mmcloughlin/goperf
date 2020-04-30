package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/mmcloughlin/cb/app/change"
	"github.com/mmcloughlin/cb/app/trace"
	"github.com/mmcloughlin/cb/pkg/command"
)

func main() {
	command.RunError(run)
}

var (
	input  = flag.String("traces", "", "trace points file")
	output = flag.String("report", "", "report directory")
)

func run(ctx context.Context, l *zap.Logger) (err error) {
	flag.Parse()
	log := l.Sugar()
	// Read trace.
	ps, err := trace.ReadPointsFile(*input)
	if err != nil {
		return err
	}

	log.Infow("read points", "num_points", len(ps))

	// Convert to traces.
	traces := trace.Traces(ps)

	log.Infow("converted to traces", "num_traces", len(traces))

	// Report.
	if err := report(traces, *output, l); err != nil {
		return err
	}

	return nil
}

func report(traces map[trace.ID]*trace.Trace, dir string, l *zap.Logger) error {
	// Ensure directory.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Change detector.
	detector := &change.Hybrid{
		WindowSize:    30,
		MinEffectSize: 4,

		M:                15,
		K:                3,
		PercentThreshold: 4,
		Context:          3,
	}

	// Build report.
	buf := bytes.NewBuffer(nil)
	n := 0
	for id, trc := range traces {
		// Look for changes.
		pts := detector.Detect(trc.Series)
		if len(pts) == 0 {
			continue
		}

		l.Debug("completed change search", zap.Stringer("id", id), zap.Int("num_changes", len(pts)))

		n++
		if n == 100 {
			l.Info("reached max number of traces")
			break
		}

		// Output.
		fmt.Fprintf(buf, "<h2>%s</h2>\n", id)
		link := benchLink(id)
		fmt.Fprintf(buf, "<p><a href=\"%s\">%s</a></p>\n", link, link)

		// Testdata.
		testdatabase := strings.ReplaceAll(trc.ID.String(), "/", "_") + ".json"
		testdatapath := filepath.Join(dir, testdatabase)
		if err := writeTestCase(testdatapath, trc.Series); err != nil {
			return err
		}

		fmt.Fprintf(buf, "<p><a href=\"%s\"><code>%s</code></a></p>\n", testdatabase, testdatabase)

		// Full plot.
		plotbase := fmt.Sprintf("trace%d.png", n)
		plotpath := filepath.Join(dir, plotbase)
		if err := plotTrace(plotpath, trc); err != nil {
			return err
		}

		fmt.Fprintf(buf, "<p><img src=\"%s\" /></p>\n", plotbase)

		// Change points.
		for _, pt := range pts {
			fmt.Fprintf(buf, "<h3>idx=%d effect=%v</h3>\n", pt.CommitIndex, pt.EffectSize)

			chgbase := fmt.Sprintf("trace%d-change%d.png", n, pt.CommitIndex)
			chgpath := filepath.Join(dir, chgbase)
			if err := plotChange(chgpath, trc, pt); err != nil {
				return err
			}
			fmt.Fprintf(buf, "<p><img src=\"%s\" /></p>\n", chgbase)
		}
	}

	// Write HTML.
	if err := ioutil.WriteFile(filepath.Join(dir, "index.html"), buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func plotTrace(filename string, t *trace.Trace) error {
	return plotSeries(filename, t.ID.String(), t.Series)
}

func plotChange(filename string, t *trace.Trace, pt change.Change) error {
	around := 20
	window := []trace.IndexedValue{}
	for _, v := range t.Series {
		if intabs(v.CommitIndex-pt.CommitIndex) < around {
			window = append(window, v)
		}
	}
	title := fmt.Sprintf("change %d effect %v", pt.CommitIndex, pt.EffectSize)
	return plotSeries(filename, title, window)
}

func plotSeries(filename string, title string, s []trace.IndexedValue) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = title
	p.X.Label.Text = "commit index"
	p.Y.Label.Text = "value"

	pts := make(plotter.XYs, len(s))
	for i, v := range s {
		pts[i].X = float64(v.CommitIndex)
		pts[i].Y = v.Value
	}

	err = plotutil.AddLinePoints(p, "series", pts)
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return err
	}

	return nil
}

func benchLink(id trace.ID) string {
	return fmt.Sprintf("https://goperf.org/bench/%s", id.BenchmarkUUID)
}

type TestCase struct {
	Expect []int        `json:"expect"`
	Series trace.Series `json:"series"`
}

func writeTestCase(filename string, s trace.Series) error {
	tc := TestCase{Expect: []int{}, Series: s}
	b, err := json.MarshalIndent(tc, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0644)
}

func intabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
