package main

import (
	"fmt"
	"math"
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
	prch := totalPRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(ch[i])
		prpts[i].X = float64(i)
		prpts[i].Y = float64(prch[i])
	}

	p.Title.Text = "Total Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
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
	prch := openPRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(ch[i])
		prpts[i].X = float64(i)
		prpts[i].Y = float64(prch[i])
	}

	p.Title.Text = "Open Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
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
		if totalh[i] == 0 {
			continue
		}
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

	p.Title.Text = "Open:Total Issues"
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

	p.Title.Text = "Age of Open Issues"
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

func drawIssueSolvedDurationOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(WeekDuration)
	qh := issueResolvedDurationQuantileHistory(issues, start, end, 0.50, 0.90, 0.99)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p25_pts := make(plotter.XYs, len(qh))
	p50_pts := make(plotter.XYs, len(qh))
	p75_pts := make(plotter.XYs, len(qh))
	for i, q := range qh {
		if q.Query(0.50) != 0 {
			p25_pts[i].X, p25_pts[i].Y = float64(i), math.Log2(q.Query(0.50))
		} else {
			p25_pts[i].X, p25_pts[i].Y = float64(i), 0
		}
		if q.Query(0.90) != 0 {
			p50_pts[i].X, p50_pts[i].Y = float64(i), math.Log2(q.Query(0.90))
		} else {
			p50_pts[i].X, p50_pts[i].Y = float64(i), 0
		}
		if q.Query(0.99) != 0 {
			p75_pts[i].X, p75_pts[i].Y = float64(i), math.Log2(q.Query(0.99))
		} else {
			p75_pts[i].X, p75_pts[i].Y = float64(i), 0
		}
	}

	p.Title.Text = "Solved Duration of Issues"
	p.X.Label.Text = fmt.Sprintf("Week from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Log2 Duration (days)"
	err = plotutil.AddLines(p, "Median", p25_pts, "90th percentile", p50_pts, "99th percentile", p75_pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawNewIssuesOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(WeekDuration)
	end := time.Now().Truncate(DayDuration).Add(WeekDuration)
	ch := newIssuesCountHistory(issues, start, end)
	prch := newPRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(ch[i])
		prpts[i].X = float64(i)
		prpts[i].Y = float64(prch[i])
	}

	p.Title.Text = "New Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Week from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawNewExternalIssuesOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(WeekDuration)
	end := time.Now().Truncate(DayDuration).Add(WeekDuration)
	ch := newExternalIssuesCountHistory(issues, start, end)
	prch := newExternalPRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(ch[i])
		prpts[i].X = float64(i)
		prpts[i].Y = float64(prch[i])
	}

	p.Title.Text = "New etcd External Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Week from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawCloseRateOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(WeekDuration)
	end := time.Now().Truncate(DayDuration).Add(WeekDuration)
	ch := newIssuesCountHistory(issues, start, end)
	prch := newPRCountHistory(issues, start, end)
	closech := closeIssuesCountHistory(issues, start, end)
	closeprch := closePRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		if closech[i] != 0 && ch[i] != 0 {
			pts[i].Y = math.Log2(float64(closech[i]) / float64(ch[i]))
		}
		prpts[i].X = float64(i)
		if closeprch[i] != 0 && prch[i] != 0 {
			prpts[i].Y = math.Log2(float64(closeprch[i]) / float64(prch[i]))
		}
	}

	p.Title.Text = "Close:New Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Week from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Log2 Fraction"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}

}

func drawOpenExternalIssuesOnDate(filename string, issues []github.Issue) {
	start := firstCreate(issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)
	ch := openExternalIssuesCountHistory(issues, start, end)
	prch := openExternalPRCountHistory(issues, start, end)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pts := make(plotter.XYs, len(ch))
	prpts := make(plotter.XYs, len(ch))
	for i := range ch {
		pts[i].X = float64(i)
		pts[i].Y = float64(ch[i])
		prpts[i].X = float64(i)
		prpts[i].Y = float64(prch[i])
	}

	p.Title.Text = "Open External Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day from %s to %s", start.Format(DateFormat), end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", pts, "PRs", prpts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawRelease(filename string, client *github.Client, owner, repo string) {
	//	releases, _, err := client.Repositories.ListReleases(owner, repo, nil)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
}

// Returns count history on total issues per day from the given start
// to the given end.
func totalIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		c := i.CreatedAt
		for k := c.Sub(start) / DayDuration; k < end.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
}

func totalPRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
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
		if i.PullRequestLinks != nil {
			continue
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
}

func openPRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
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
		if i.PullRequestLinks != nil {
			continue
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}

		firsti := created.Sub(start) / DayDuration
		for k := firsti; k < closed.Sub(start)/DayDuration; k++ {
			qs[k].Insert(float64(k - firsti))
		}
	}
	return qs
}

func issueResolvedDurationQuantileHistory(issues []github.Issue, start, end time.Time, quantileHints ...float64) []*quantile.Stream {
	qs := make([]*quantile.Stream, end.Sub(start)/WeekDuration)
	for i := range qs {
		if len(quantileHints) != 0 {
			qs[i] = quantile.NewTargeted(quantileHints...)
		} else {
			qs[i] = quantile.NewBiased()
		}
	}
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		d := end.Sub(start)
		if i.ClosedAt != nil {
			d = i.ClosedAt.Sub(*i.CreatedAt)
		}
		qs[i.CreatedAt.Sub(start)/WeekDuration].Insert(float64(d) / float64(DayDuration))
	}
	return qs
}

func newIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		counts[i.CreatedAt.Sub(start)/WeekDuration]++
	}
	return counts
}

func newPRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
		counts[i.CreatedAt.Sub(start)/WeekDuration]++
	}
	return counts
}

func newExternalIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		name := *i.User.Login
		if name != "xiang90" && name != "yichengq" && name != "philips" && name != "bcwaldon" && name != "jonboulle" && name != "kelseyhightower" && name != "bmizerany" && name != "barakmich" && name != "robszumski" {
			counts[i.CreatedAt.Sub(start)/WeekDuration]++
		}
	}
	return counts
}

func newExternalPRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
		name := *i.User.Login
		if name != "xiang90" && name != "yichengq" && name != "philips" && name != "bcwaldon" && name != "jonboulle" && name != "kelseyhightower" && name != "bmizerany" && name != "barakmich" && name != "robszumski" {
			counts[i.CreatedAt.Sub(start)/WeekDuration]++
		}
	}
	return counts
}

func closeIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		if i.ClosedAt != nil {
			counts[i.ClosedAt.Sub(start)/WeekDuration]++
		}
	}
	return counts
}

func closePRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/WeekDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
		if i.ClosedAt != nil {
			counts[i.ClosedAt.Sub(start)/WeekDuration]++
		}
	}
	return counts
}

func openExternalIssuesCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		if i.PullRequestLinks != nil {
			continue
		}
		name := *i.User.Login
		if name != "xiang90" && name != "yichengq" && name != "philips" && name != "bcwaldon" && name != "jonboulle" && name != "kelseyhightower" && name != "bmizerany" && name != "barakmich" && name != "robszumski" {
			continue
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
}

func openExternalPRCountHistory(issues []github.Issue, start, end time.Time) []int {
	counts := make([]int, end.Sub(start)/DayDuration)
	for _, i := range issues {
		if i.PullRequestLinks == nil {
			continue
		}
		name := *i.User.Login
		if name != "xiang90" && name != "yichengq" && name != "philips" && name != "bcwaldon" && name != "jonboulle" && name != "kelseyhightower" && name != "bmizerany" && name != "barakmich" && name != "robszumski" {
			continue
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
			counts[k]++
		}
	}
	return counts
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
