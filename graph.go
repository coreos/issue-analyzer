package main

import (
	"fmt"
	"time"

	"github.com/bmizerany/perks/quantile"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotutil"
	"github.com/google/go-github/github"
)

func drawTotalIssues(ctx *context, filename string) {
	start := firstCreate(ctx.issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)

	issues := make([]int, end.Sub(start)/DayDuration)
	prs := make([]int, end.Sub(start)/DayDuration)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		c := i.CreatedAt
		for k := c.Sub(start) / DayDuration; k < end.Sub(start)/DayDuration; k++ {
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
	start := firstCreate(ctx.issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)

	issues := make([]int, end.Sub(start)/DayDuration)
	prs := make([]int, end.Sub(start)/DayDuration)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
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
	start := firstCreate(ctx.issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)

	totals := make([]int, end.Sub(start)/DayDuration)
	opens := make([]int, end.Sub(start)/DayDuration)
	ctx.WalkIssues(func(i github.Issue, isPullRequest bool) {
		if isPullRequest {
			return
		}
		created := i.CreatedAt
		closed := end
		if i.ClosedAt != nil {
			closed = *i.ClosedAt
		}
		for k := created.Sub(start) / DayDuration; k < end.Sub(start)/DayDuration; k++ {
			totals[k]++
		}
		for k := created.Sub(start) / DayDuration; k < closed.Sub(start)/DayDuration; k++ {
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
	start := firstCreate(ctx.issues).Truncate(DayDuration)
	end := time.Now().Truncate(DayDuration).Add(DayDuration)

	qs := make([]*quantile.Stream, end.Sub(start)/DayDuration)
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
		for k := firsti; k < closed.Sub(start)/DayDuration; k++ {
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
	start := firstCreate(ctx.issues).Truncate(MonthDuration)
	end := time.Now().Truncate(DayDuration).Add(MonthDuration)

	qs := make([]*quantile.Stream, end.Sub(start)/MonthDuration)
	for i := range qs {
		qs[i] = quantile.NewTargeted(0.25, 0.50, 0.75)
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
		for k := i.CreatedAt.Sub(start) / MonthDuration; k < end.Sub(start)/MonthDuration; k++ {
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
	err = plotutil.AddLines(p, "25th percentile", quantileAt(qs, 0.25), "50th percentile", quantileAt(qs, 0.50), "75th percentile", quantileAt(qs, 0.75))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

/*
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
*/

func firstCreate(issues []github.Issue) time.Time {
	first := time.Now()
	for _, i := range issues {
		if i.CreatedAt.Before(first) {
			first = *i.CreatedAt
		}
	}
	return first
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
