package main

import (
	"fmt"
	"time"

	"github.com/bmizerany/perks/quantile"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/google/go-github/github"
)

func drawTotalIssuesOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)
	ch := totalIssuesCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	for i, c := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(c)
	}

	p.Title.Text = "Total Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssuesOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)
	ch := openIssuesCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	for i, c := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(c)
	}

	p.Title.Text = "Open Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueFractionOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)
	totalh := totalIssuesCountHistory(issues, start, end)
	openh := openIssuesCountHistory(issues, start, end)

	fractionh := make([]float64, len(totalh))
	for i := range totalh {
		fractionh[i] = float64(openh[i]) / float64(totalh[i])
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(fractionh))
	for i, f := range fractionh {
		pts[i].X = float64(i)
		pts[i].Y = f
	}

	p.Title.Text = "Open:Total Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Fraction"
	err = plotutil.AddLines(p, pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueAgeOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)
	qh := openDaysQuantileHistory(issues, start, end, 0.25, 0.50, 0.75)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p25_pts := make(plotter.XYs, len(qh))
	p50_pts := make(plotter.XYs, len(qh))
	p75_pts := make(plotter.XYs, len(qh))
	for i, q := range qh {
		p25_pts[i].X, p25_pts[i].Y = float64(i), q.Query(0.25)
		p50_pts[i].X, p50_pts[i].Y = float64(i), q.Query(0.50)
		p75_pts[i].X, p75_pts[i].Y = float64(i), q.Query(0.75)
	}

	p.Title.Text = "Age of Open Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Age (days)"
	err = plotutil.AddLines(p, "25th percentile", p25_pts, "Median", p50_pts, "75th percentile", p75_pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

// Returns count history on total issues per day from the given start
// to the given end.
func totalIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		c := i.CreatedAt
		for k := c.Sub(start) / DayDuration; k < end.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
}

func openIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = i.ClosedAt.Add(DayDuration)
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
}

// quantileHints helps to calculate quantile values with less resources and finer error guarantees.
func openDaysQuantileHistory(issues []github.Issue, start, end time.Time, quantileHints ...float64) []*quantile.Stream {
	qs := make([]*quantile.Stream, end.Sub(start)/DayDuration)
	for i := range qs {
		if len(quantileHints) != 0 {
			qs[i] = quantile.NewTargeted(quantileHints...)
		} else {
			qs[i] = quantile.NewBiased()
		}
	}
	for _, i := range issues {
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = i.ClosedAt.Add(DayDuration)
		}

		firsti := created.Sub(start) / DayDuration
		for k := firsti; k < closed.Sub(start)/DayDuration; k++ {
			qs[k].Insert(float64(k - firsti))
		}
	}
	return qs
}

func firstCreate(issues []github.Issue) time.Time {
	first := time.Now()
	for _, i := range issues {
		if i.CreatedAt.Before(first) {
			first = *i.CreatedAt
		}
	}
	return first
}
