package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gonum/plot/vg"
)

const (
	DayDuration   = 24 * time.Hour
	WeekDuration  = 7 * DayDuration
	MonthDuration = 30 * DayDuration
	DateFormat    = "2006-01-02"

	defaultWidth  = 6 * vg.Inch
	defaultHeight = 4 * vg.Inch
)

func main() {
	owner := flag.String("owner", "coreos", "the owner in github")
	repo := flag.String("repo", "etcd", "the repo of the owner in github")
	token := flag.String("token", "", "access token for github")
	start := flag.String("start-date", "", "start date of the graph, in format 2000-Jan-01 or 2000-Jan")
	end := flag.String("end-date", "", "end date of the graph, in format 2000-Jan-01 or 2000-Jan")
	flag.Parse()

	if *token == "" {
		if data, err := ioutil.ReadFile(".oauth2_token"); err == nil {
			*token = string(data)
		}
	}
	if *token == "" {
		fmt.Println("Using unauthenticated client because oauth2 token is unavailable,")
		fmt.Println("whose rate is limited to 60 requests per hour.")
		fmt.Println("Learn more about GitHub rate limiting at http://developer.github.com/v3/#rate-limiting.")
		fmt.Println("If you want to use authenticated client, please save your oauth token into file './.oauth2_token'.")
	} else {
		fmt.Println("Using authenticated client whose rate is up to 5000 requests per hour.")
	}

	rc := newRepoClient(*owner, *repo, *token)

	rc.LoadIssues()
	rc.LoadReleases()
	per := newPeriod(rc, parseDateString(*start), parseDateString(*end))

	drawTotalIssues(rc, per, "total_issues.png")
	drawOpenIssues(rc, per, "open_issues.png")
	drawOpenIssueFraction(rc, per, "open_fraction.png")
	drawOpenIssueAge(rc, per, "open_age.png")
	drawIssueSolvedDuration(rc, per, "solved_duration.png")
	drawTopReleaseDownloads(rc, per, "top_downloads.png")
	buildImagesHTML("images.html", "total_issues.png", "open_issues.png", "open_fraction.png", "open_age.png", "solved_duration.png", "top_downloads.png")
	fmt.Printf("saved images and browsing html\n")

	startBrowser("images.html")
}

func parseDateString(date string) time.Time {
	if date == "" {
		return time.Time{}
	}
	if t, err := time.Parse("2006-Jan-02", date); err == nil {
		return t
	}
	if t, err := time.Parse("2006-Jan-02", fmt.Sprint(date, "-01")); err == nil {
		return t
	}
	fmt.Fprintf(os.Stderr, "malformat date string %q\n", date)
	os.Exit(1)
	return time.Time{}
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
