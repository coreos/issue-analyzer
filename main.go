package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bmizerany/perks/quantile"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	DayDuration = 24 * time.Hour
	DateFormat  = "2006-01-02"

	issueCacheFilename = "issues.cache"
)

func main() {
	owner := flag.String("owner", "coreos", "the owner in github")
	repo := flag.String("repo", "etcd", "the repo of the owner in github")
	token := flag.String("token", "", "access token for github")
	flag.Parse()

	if *token == "" {
		if data, err := ioutil.ReadFile(".oauth2_token"); err == nil {
			*token = string(data)
		}
	}

	var c *http.Client
	if *token == "" {
		fmt.Println("Using unauthenticated client because oauth2 token is unavailable,")
		fmt.Println("whose rate is limited to 60 requests per hour.")
		fmt.Println("Learn more about GitHub rate limiting at http://developer.github.com/v3/#rate-limiting.")
		fmt.Println("If you want to use authenticated client, please save your oauth token into file './.oauth2_token'.")
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		c = oauth2.NewClient(oauth2.NoContext, ts)
		fmt.Println("Using authenticated client whose rate is up to 5000 requests per hour.")
	}
	client := github.NewClient(c)

	var issues []github.Issue
	data, err := ioutil.ReadFile(issueCacheFilename)
	haveCachedIssue := err != nil
	isUpToDate := time.Now().Sub(fileModTime(issueCacheFilename)) < DayDuration
	if haveCachedIssue && isUpToDate {
		issues = allIssuesInRepo(client, *owner, *repo)
		if data, err := json.Marshal(issues); err != nil {
			fmt.Printf("error marshaling issues (%v)\n", err)
		} else if err := ioutil.WriteFile(issueCacheFilename, data, 0600); err != nil {
			fmt.Printf("error caching issues into file (%v)\n", err)
		} else {
			fmt.Printf("cached issues in file %q for fast retrieval\n", issueCacheFilename)
		}
	} else {
		if err := json.Unmarshal(data, &issues); err != nil {
			fmt.Printf("error loading cached issues (%v)\n", err)
			fmt.Printf("Please remove file %s and run the command again.", issueCacheFilename)
			os.Exit(1)
		}
	}

	drawTotalIssuesOnDate("total_issues.png", issues)
	drawOpenIssuesOnDate("open_issues.png", issues)
	drawOpenIssueFractionOnDate("open_fraction.png", issues)
	drawOpenIssueAgeOnDate("open_age.png", issues)
}

func allIssuesInRepo(client *github.Client, owner, repo string) []github.Issue {
	rate, _, err := client.RateLimits()
	if err != nil {
		fmt.Printf("error fetching rate limit (%v)\n", err)
	} else {
		fmt.Printf("API Rate Limit: %s\n", rate)
	}

	opt := &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	var issues []github.Issue
	for i := 0; ; i++ {
		is, resp, err := client.Issues.ListByRepo(owner, repo, opt)
		if err != nil {
			fmt.Printf("error listing issues (%v)\n", err)
			os.Exit(1)
		}
		issues = append(issues, is...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
		fmt.Printf("list %d issues...\n", len(issues))
	}
	return issues
}

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
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
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
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
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
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
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
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
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

func fileModTime(name string) time.Time {
	f, err := os.Open(name)
	if err != nil {
		return time.Time{}
	}
	st, err := f.Stat()
	if err != nil {
		return time.Time{}
	}
	return st.ModTime()
}
