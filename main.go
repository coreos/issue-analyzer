package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gonum/plot/vg"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	DayDuration = 24 * time.Hour
	DateFormat  = "2006-01-02"

	defaultWidth  = 6 * vg.Inch
	defaultHeight = 4 * vg.Inch
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
	issueCacheFilename := fmt.Sprintf("cache/%s_%s_issues.cache", *owner, *repo)
	data, err := ioutil.ReadFile(issueCacheFilename)
	haveCachedIssue := err == nil
	isUpToDate := time.Now().Sub(fileModTime(issueCacheFilename)) < DayDuration
	if haveCachedIssue && isUpToDate {
		if err := json.Unmarshal(data, &issues); err != nil {
			fmt.Printf("error loading cached issues (%v)\n", err)
			fmt.Printf("Please remove file %s and run the command again.", issueCacheFilename)
			os.Exit(1)
		}
	} else {
		issues = allIssuesInRepo(client, *owner, *repo)
		if data, err := json.Marshal(issues); err != nil {
			fmt.Printf("error marshaling issues (%v)\n", err)
		} else if err := ioutil.WriteFile(issueCacheFilename, data, 0600); err != nil {
			fmt.Printf("error caching issues into file (%v)\n", err)
		} else {
			fmt.Printf("cached issues in file %q for fast retrieval\n", issueCacheFilename)
		}
	}

	drawTotalIssuesOnDate("total_issues.png", issues)
	drawOpenIssuesOnDate("open_issues.png", issues)
	drawOpenIssueFractionOnDate("open_fraction.png", issues)
	drawOpenIssueAgeOnDate("open_age.png", issues)
	buildImagesHTML("images.html", "total_issues.png", "open_issues.png", "open_fraction.png", "open_age.png")
	fmt.Printf("saved images and browsing html\n")

	startBrowser("images.html")
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

func buildImagesHTML(html string, images ...string) {
	var body string
	for _, i := range images {
		body = body + fmt.Sprintf("<img src=%q>\n", i)
	}
	err := ioutil.WriteFile(html, []byte(body), 0666)
	if err != nil {
		panic(err)
	}
}

func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}
