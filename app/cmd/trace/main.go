package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"go.skia.org/infra/go/vec32"
	"go.skia.org/infra/perf/go/dataframe"
	"go.skia.org/infra/perf/go/types"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"go.skia.org/infra/perf/go/clustering2"
	"go.skia.org/infra/perf/go/config"

	"github.com/google/uuid"
	"github.com/mmcloughlin/cb/app/trace"

	"github.com/mmcloughlin/cb/pkg/command"
	"go.uber.org/zap"
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

	// Ensure directory.
	if err := os.MkdirAll(*output, 0755); err != nil {
		return err
	}

	// Read trace.
	ps, err := trace.ReadPointsFile(*input)
	if err != nil {
		return err
	}

	l.Info("read traces", zap.Int("num_points", len(ps)))

	// Convert to dataframe.
	df := trace.PointsDataFrame(ps)

	l.Info("built dataframe",
		zap.Int("num_traces", len(df.TraceSet)),
		zap.Int("num_columns", len(df.Header)),
	)

	if err := plots(df, *output); err != nil {
		return err
	}

	// Clustering.
	summaries, err := clustering2.CalculateClusterSummaries(
		df,
		100, // k
		config.MinStdDev,
		nil,
		1.0, // interesting
		types.CohenStep,
	)
	if err != nil {
		return err
	}

	l.Info("calculated cluster summaries", zap.Int("num_clusters", len(summaries.Clusters)))

	// Generate report.
	if err := report(summaries, *output); err != nil {
		return err
	}

	return nil
}

func plots(df *dataframe.DataFrame, dir string) error {
	buf := bytes.NewBuffer(nil)

	n := 0
	for key, trace := range df.TraceSet {
		if n++; n == 100 {
			break
		}

		p, err := plot.New()
		if err != nil {
			return err
		}

		p.Title.Text = key
		p.X.Label.Text = "commit index"
		p.Y.Label.Text = "value"

		line := vec32.Dup(trace)
		vec32.Fill(line)
		err = plotutil.AddLinePoints(p, "line", plot32(line))
		if err != nil {
			return err
		}

		// Save the plot to a PNG file.
		plotbase := "plot" + key + ".png"
		plotpath := filepath.Join(dir, plotbase)
		if err := p.Save(4*vg.Inch, 4*vg.Inch, plotpath); err != nil {
			return err
		}

		fmt.Fprintf(buf, "<p><a href=\"%s\"><img src=\"%s\" /><a></p>\n", benchLink(key), plotbase)
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "plots.html"), buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func report(summaries *clustering2.ClusterSummaries, dir string) error {
	// Build report.
	buf := bytes.NewBuffer(nil)
	for i, cluster := range summaries.Clusters {
		name := "cluster" + strconv.Itoa(i)
		if err := clusterReport(buf, name, cluster, dir); err != nil {
			return err
		}
	}

	// Write HTML.
	if err := ioutil.WriteFile(filepath.Join(dir, "index.html"), buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func clusterReport(buf *bytes.Buffer, name string, summary *clustering2.ClusterSummary, dir string) error {
	// heading
	fmt.Fprintf(buf, "<h2>%s: %s</h2>\n", summary.StepFit.Status, name)

	// plot
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = name
	p.X.Label.Text = "commit index"
	p.Y.Label.Text = "value"

	err = plotutil.AddLinePoints(p, "First", plot32(summary.Centroid))
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	plotbase := name + ".png"
	plotpath := filepath.Join(dir, plotbase)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, plotpath); err != nil {
		return err
	}

	fmt.Fprintf(buf, "<img src=\"%s\" />\n", plotbase)

	// keys
	fmt.Fprint(buf, "<pre>\n")
	for _, k := range summary.Keys {
		fmt.Fprintf(buf, "<a href=\"%s\">%s</a>\n", benchLink(k), k)
	}
	fmt.Fprint(buf, "</pre>\n")

	return nil
}

func plot32(ys []float32) plotter.XYs {
	pts := make(plotter.XYs, len(ys))
	for i, y := range ys {
		pts[i].X = float64(i)
		pts[i].Y = float64(y)
	}
	return pts
}

func benchLink(key string) string {
	prefix := len(",benchmark_uuid=")
	benchID := key[prefix : prefix+32]
	benchUUID := uuid.MustParse(benchID)
	return fmt.Sprintf("https://goperf.org/bench/%s", benchUUID)
}
