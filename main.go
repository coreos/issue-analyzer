package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

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
)

func main() {
	var token string
	if data, err := ioutil.ReadFile(".oauth2_token"); err == nil {
		token = string(data)
	}

	var c *http.Client
	if token == "" {
		fmt.Println("Using unauthenticated client because oauth2 token is unavailable,")
		fmt.Println("whose rate is limited to 60 requests per hour.")
		fmt.Println("Learn more about GitHub rate limiting at http://developer.github.com/v3/#rate-limiting.")
		fmt.Println("If you want to use authenticated client, please save your oauth token into file './.oauth2_token'.")
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		c = oauth2.NewClient(oauth2.NoContext, ts)
		fmt.Println("Using authenticated client whose rate is up to 5000 requeests per hour.")
	}
	client := github.NewClient(c)

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
		is, resp, err := client.Issues.ListByRepo("coreos", "etcd", opt)
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

	drawTotalIssuesOnDate("total_issues.png", issues)
	drawOpenIssuesOnDate("open_issues.png", issues)
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

func firstCreate(issues []github.Issue) time.Time {
	first := time.Now()
	for _, i := range issues {
		if i.CreatedAt.Before(first) {
			first = *i.CreatedAt
		}
	}
	return first
}
