package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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
	creates := make(map[time.Time]int)
	for _, i := range issues {
		c := (*i.CreatedAt).Truncate(24 * time.Hour)
		creates[c]++
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	xs := make([]time.Time, 0, len(creates))
	for t := range creates {
		xs = append(xs, t)
	}
	sort.Sort(timeSlice(xs))

	start := xs[0]
	pts := make(plotter.XYs, len(creates))
	var prev int
	for i, x := range xs {
		pts[i].X = float64(x.Sub(start) / 24 / time.Hour)
		pts[i].Y = float64(prev + creates[x])
		prev += creates[x]
	}

	p.Title.Text = "Total Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day since %s", start.String())
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
	today := time.Now().Truncate(24 * time.Hour)
	openm := make(map[time.Time]int)
	for _, i := range issues {
		create := i.CreatedAt.Truncate(24 * time.Hour)
		last := today
		if i.ClosedAt != nil {
			last = i.ClosedAt.Truncate(24 * time.Hour)
		}
		for k := create; !k.After(last); k = k.Add(24 * time.Hour) {
			openm[k]++
		}
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	xs := make([]time.Time, 0, len(openm))
	for t := range openm {
		xs = append(xs, t)
	}
	sort.Sort(timeSlice(xs))

	start := xs[0]
	pts := make(plotter.XYs, len(openm))
	for i, x := range xs {
		pts[i].X = float64(x.Sub(start) / 24 / time.Hour)
		pts[i].Y = float64(openm[x])
	}

	p.Title.Text = "Open Issues/PR"
	p.X.Label.Text = fmt.Sprintf("Day since %s", start.String())
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

type timeSlice []time.Time

func (s timeSlice) Len() int           { return len(s) }
func (s timeSlice) Less(i, j int) bool { return s[i].Before(s[j]) }
func (s timeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
