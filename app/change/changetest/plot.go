package changetest

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/mmcloughlin/cb/app/trace"
)

// PlotSeries generates a plot for debugging purposes.
func PlotSeries(filename, title string, s trace.Series) error {
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

	// Save the plot.
	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return err
	}

	return nil
}
