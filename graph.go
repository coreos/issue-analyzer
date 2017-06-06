package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/bmizerany/perks/quantile"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/google/go-github/github"
)

type period struct {
	rc    *repoClient
	start time.Time
	end   time.Time
}

func newPeriod(rc *repoClient, start, end time.Time) *period {
	p := &period{rc, start, end}
	if start.IsZero() || rc.StartTime().After(start) {
		p.start = rc.StartTime()
	}
	if end.IsZero() || rc.EndTime().Before(end) {
		p.end = rc.EndTime()
	}
	return p
}

func (p *period) seqInts(a []int, interval time.Duration) seqInts {
	i := p.start.Sub(p.rc.StartTime()) / interval
	j := p.end.Sub(p.rc.StartTime()) / interval
	return a[i:j]
}

func (p *period) seqFloats(a []float64, interval time.Duration) seqFloats {
	i := p.start.Sub(p.rc.StartTime()) / interval
	j := p.end.Sub(p.rc.StartTime()) / interval
	return a[i:j]
}

func drawTotalIssues(rc *repoClient, per *period, filename string) {
	start, end := rc.StartTime(), rc.EndTime()

	l := end.Sub(start)/DayDuration + 1
	issues := make([]int, l)
	prs := make([]int, l)
	rc.WalkIssues(func(i github.Issue, isPullRequest bool) {
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
	p.X.Label.Text = fmt.Sprintf("Date from %s to %s", per.start.Format(DateFormat), per.end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", per.seqInts(issues, DayDuration), "PRs", per.seqInts(prs, DayDuration))
	if err != nil {
		panic(err)
	}
	p.X.Tick.Marker = newDayTicker(p.X.Tick.Marker, per.start)

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssues(rc *repoClient, per *period, filename string) {
	start, end := rc.StartTime(), rc.EndTime()

	l := end.Sub(start)/DayDuration + 1
	issues := make([]int, l)
	prs := make([]int, l)
	rc.WalkIssues(func(i github.Issue, isPullRequest bool) {
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
	p.X.Label.Text = fmt.Sprintf("Date from %s to %s", per.start.Format(DateFormat), per.end.Format(DateFormat))
	p.Y.Label.Text = "Count"
	err = plotutil.AddLines(p, "issues", per.seqInts(issues, DayDuration), "PRs", per.seqInts(prs, DayDuration))
	if err != nil {
		panic(err)
	}
	p.X.Tick.Marker = newDayTicker(p.X.Tick.Marker, per.start)

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueFraction(rc *repoClient, per *period, filename string) {
	start, end := rc.StartTime(), rc.EndTime()

	l := end.Sub(start)/DayDuration + 1
	totals := make([]int, l)
	opens := make([]int, l)
	rc.WalkIssues(func(i github.Issue, isPullRequest bool) {
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
	p.X.Label.Text = fmt.Sprintf("Date from %s to %s", per.start.Format(DateFormat), per.end.Format(DateFormat))
	p.Y.Label.Text = "Fraction"
	err = plotutil.AddLines(p, per.seqFloats(fractions, DayDuration))
	if err != nil {
		panic(err)
	}
	p.X.Tick.Marker = newDayTicker(p.X.Tick.Marker, per.start)

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawOpenIssueAge(rc *repoClient, per *period, filename string) {
	start, end := rc.StartTime(), rc.EndTime()

	l := end.Sub(start)/DayDuration + 1
	qs := make([]*quantile.Stream, l)
	for i := range qs {
		qs[i] = quantile.NewTargeted(0.25, 0.50, 0.75)
	}
	rc.WalkIssues(func(i github.Issue, isPullRequest bool) {
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
	p.X.Label.Text = fmt.Sprintf("Date from %s to %s", per.start.Format(DateFormat), per.end.Format(DateFormat))
	p.Y.Label.Text = "Age (days)"
	err = plotutil.AddLines(p, "25th percentile", per.seqFloats(quantileAt(qs, 0.25), DayDuration),
		"Median", per.seqFloats(quantileAt(qs, 0.50), DayDuration),
		"75th percentile", per.seqFloats(quantileAt(qs, 0.75), DayDuration))
	if err != nil {
		panic(err)
	}
	p.X.Tick.Marker = newDayTicker(p.X.Tick.Marker, per.start)

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawIssueSolvedDuration(rc *repoClient, per *period, filename string) {
	start, end := rc.StartTime(), rc.EndTime()

	l := end.Sub(start)/MonthDuration + 1
	qs := make([]*quantile.Stream, l)
	for i := range qs {
		qs[i] = quantile.NewTargeted(0.50)
	}
	rc.WalkIssues(func(i github.Issue, isPullRequest bool) {
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
	p.X.Label.Text = fmt.Sprintf("Month from %s to %s", per.start.Format(DateFormat), per.end.Format(DateFormat))
	p.Y.Label.Text = "Duration (days)"
	err = plotutil.AddLines(p, "Median", per.seqFloats(quantileAt(qs, 0.50), MonthDuration))
	if err != nil {
		panic(err)
	}
	p.X.Tick.Marker = newMonthTicker(p.X.Tick.Marker, per.start)

	// Save the plot to a PNG file.
	if err := p.Save(defaultWidth, defaultHeight, filename); err != nil {
		panic(err)
	}
}

func drawTopReleaseDownloads(rc *repoClient, per *period, filename string) {
	var rs releases
	rc.WalkReleases(func(r github.RepositoryRelease) {
		var cnt int
		if r.CreatedAt.Before(per.start) || r.CreatedAt.After(per.end) {
			return
		}
		for _, a := range r.Assets {
			cnt += *a.DownloadCount
		}
		rs = append(rs, release{name: *r.TagName, download: cnt})
	})
	sort.Sort(rs)

	var names []string
	var downloads []int
	num := 10
	if num > len(rs) {
		num = len(rs)
	}
	for i := 0; i < num; i++ {
		names = append(names, rs[i].name)
		downloads = append(downloads, rs[i].download)
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Release Downloads"
	p.Y.Label.Text = "Download Count"
	if len(names) > 0 {
		p.NominalX(names...)
		bars, err := plotter.NewBarChart(ints(downloads), vg.Points(20))
		if err != nil {
			panic(err)
		}
		bars.LineStyle.Width = vg.Length(0)
		p.Add(bars)
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

type ints []int

func (a ints) Len() int            { return len(a) }
func (a ints) Value(i int) float64 { return float64(a[i]) }

type release struct {
	name     string
	download int
}

type releases []release

func (a releases) Len() int           { return len(a) }
func (a releases) Less(i, j int) bool { return a[i].download > a[j].download }
func (a releases) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func quantileAt(ss []*quantile.Stream, q float64) seqFloats {
	fs := make(seqFloats, len(ss))
	for i := range ss {
		fs[i] = ss[i].Query(q)
	}
	return fs
}

type dateTicker struct {
	plot.Ticker
	start    time.Time
	interval time.Duration
}

func (dt *dateTicker) Ticks(min, max float64) []plot.Tick {
	ts := dt.Ticker.Ticks(min, max)
	for i, t := range ts {
		if t.Label != "" {
			t.Label = dt.start.Add(time.Duration(t.Value) * dt.interval).Format(DateFormat)
		}
		ts[i] = t
	}
	return ts
}

func newDayTicker(t plot.Ticker, start time.Time) plot.Ticker {
	return &dateTicker{
		Ticker:   t,
		start:    start,
		interval: DayDuration,
	}
}

func newMonthTicker(t plot.Ticker, start time.Time) plot.Ticker {
	return &dateTicker{
		Ticker:   t,
		start:    start,
		interval: MonthDuration,
	}
}
