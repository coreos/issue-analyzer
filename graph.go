package main

import (
	"fmt"

	"github.com/bmizerany/perks/quantile"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotutil"
	"github.com/google/go-github/github"
)

func drawTotalIssues(ctx *context, filename string) {
	start, end := ctx.StartTime(), ctx.EndTime()

	l := end.Sub(start)/DayDuration + 1
	issues := make([]int, l)
	prs := make([]int, l)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		c := i.CreatedAt
		for k := c.Sub(start) / DayDuration; k <= end.Sub(start)/DayDuration; k++ {
			if isPullRequest {
				prs[k]++
			} else {
				issues[k]++
			}
		}
	})

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Total Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", seqInts(issues), "PRs", seqInts(prs))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssues(ctx *context, filename string) {
	start, end := ctx.StartTime(), ctx.EndTime()

	l := end.Sub(start)/DayDuration + 1
	issues := make([]int, l)
	prs := make([]int, l)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k <= closed.Sub(start)/DayDuration; k++ {
			if isPullRequest {
				prs[k]++
			} else {
				issues[k]++
			}
		}
	})

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Open Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", seqInts(issues), "PRs", seqInts(prs))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueFraction(ctx *context, filename string) {
	start, end := ctx.StartTime(), ctx.EndTime()

	l := end.Sub(start)/DayDuration + 1
	totals := make([]int, l)
	opens := make([]int, l)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		if isPullRequest {
			return
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k <= end.Sub(start)/DayDuration; k++ {
			totals[k]++
		}
		for k := created.Sub(start) / DayDuration; k <= closed.Sub(start)/DayDuration; k++ {
			opens[k]++
		}
	})

	fractions := make([]float64, len(totals))
	for i := range totals {
		if totals[i] != 0 {
			fractions[i] = float64(opens[i]) / float64(totals[i])
		}
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Open:Total Issues"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Fraction"
	err = plotutil.AddLines(p, seqFloats(fractions))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueAge(ctx *context, filename string) {
	start, end := ctx.StartTime(), ctx.EndTime()

	l := end.Sub(start)/DayDuration + 1
	qs := make([]*quantile.Stream, l)
	for i := range qs {
		qs[i] = quantile.NewTargeted(0.25, 0.50, 0.75)
	}
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		if isPullRequest {
			return
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		firsti := created.Sub(start) / DayDuration
		for k := firsti; k <= closed.Sub(start)/DayDuration; k++ {
			qs[k].Insert(float64(k - firsti))
		}
	})

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Age of Open Issues"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Age (days)"
	err = plotutil.AddLines(p, "25th percentile", quantileAt(qs, 0.25), "Median", quantileAt(qs, 0.50), "75th percentile", quantileAt(qs, 0.75))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawIssueSolvedDuration(ctx *context, filename string) {
	start, end := ctx.StartTime(), ctx.EndTime()

	l := end.Sub(start)/MonthDuration + 1
	qs := make([]*quantile.Stream, l)
	for i := range qs {
		qs[i] = quantile.NewTargeted(0.50)
	}
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		if isPullRequest {
			return
		}
		// count unresolved as the longest period
		d := end.Sub(start)
		if i.ClosedAt != nil {
			d = i.ClosedAt.Sub(*i.CreatedAt)
		}
		for k := i.CreatedAt.Sub(start) / MonthDuration; k <= end.Sub(start)/MonthDuration; k++ {
			qs[k].Insert(float64(d) / float64(DayDuration))
		}
	})

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Solved Duration of Issues"
	p.X.Label.Text = fmt.Sprintf("Month from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Duration (days)"
	err = plotutil.AddLines(p, "Median", quantileAt(qs, 0.50))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

type seqInts []int

func (xys seqInts) Len() int                { return len(xys) }
func (xys seqInts) XY(i int) (x, y float64) { return float64(i), float64(xys[i]) }

type seqFloats []float64

func (xys seqFloats) Len() int                { return len(xys) }
func (xys seqFloats) XY(i int) (x, y float64) { return float64(i), xys[i] }

func quantileAt(ss []*quantile.Stream, q float64) seqFloats {
	fs := make(seqFloats, len(ss))
	for i := range ss {
		fs[i] = ss[i].Query(q)
	}
	return fs
}
