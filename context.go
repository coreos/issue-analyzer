package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/go-github/github"
)

type context struct {
	client *github.Client
	owner  string
	repo   string

	issues   []github.Issue
	releases []github.RepositoryRelease
}

func (c *context) LoadIssues() {
	issueCacheFilename := fmt.Sprintf("cache/%s_%s_issues.cache", c.owner, c.repo)
	if err := readJson(issueCacheFilename, &c.issues); err == nil {
		return
	}
	c.issues = allIssuesInRepo(c.client, c.owner, c.repo)
	writeJson(issueCacheFilename, c.issues)
}

func (c *context) LoadReleases() {
	cacheFilename := fmt.Sprintf("cache/%s_%s_releases.cache", c.owner, c.repo)
	if err := readJson(cacheFilename, &c.releases); err == nil {
		return
	}
	c.releases = c.fetchReleases()
	writeJson(cacheFilename, c.releases)
}

func (c *context) StartTime() time.Time {
	first := time.Now()
	for _, i := range c.issues {
		if i.CreatedAt.Before(first) {
			first = *i.CreatedAt
		}
	}
	return first
}

func (c *context) EndTime() time.Time { return time.Now() }

func (c *context) WalkIssues(f func(issue github.Issue, isPullRequest bool)) {
	for _, issue := range c.issues {
		f(issue, issue.PullRequestLinks != nil)
	}
}

func (c *context) fetchReleases() []github.RepositoryRelease {
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var releases []github.RepositoryRelease
	for i := 0; ; i++ {
		rs, resp, err := c.client.Repositories.ListReleases(c.owner, c.repo, opt)
		if err != nil {
			fmt.Printf("error listing releases (%v)\n", err)
			os.Exit(1)
		}
		releases = append(releases, rs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return releases
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

func readJson(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	haveCachedIssue := err == nil
	isUpToDate := time.Now().Sub(fileModTime(filename)) < DayDuration
	if haveCachedIssue && isUpToDate {
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("error loading cached data from file %s (%v)", filename, err)
		}
		return nil
	}
	return fmt.Errorf("outdated cache file")
}

func writeJson(filename string, v interface{}) {
	if data, err := json.Marshal(v); err != nil {
		fmt.Printf("error marshaling issues (%v)\n", err)
	} else if err := ioutil.WriteFile(filename, data, 0600); err != nil {
		fmt.Printf("error caching issues into file (%v)\n", err)
	} else {
		fmt.Printf("cached issues in file %q for fast retrieval\n", filename)
	}
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
