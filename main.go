package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/becoded/go-harvest/harvest"
	"github.com/google/go-github/github"
	"github.com/hashicorp/hcl"
	"golang.org/x/oauth2"
)

func main() {
	harvestID := os.Getenv("HARVEST_ACCOUNT_ID")
	harvestToken := os.Getenv("HARVEST_ACCESS_TOKEN")
	timeEntries, err := FetchTimeEntries(harvestID, harvestToken)
	if err != nil {
		log.Fatal(err)
	}

	timeEntryDates := make(map[string]bool)
	for _, timeEntry := range timeEntries {
		timeEntryDates[timeEntry.SpentDate.Format("2006-01-02")] = true
	}

	projectMappings, err := readProjectMapping("harvespex.hcl")
	if err != nil {
		log.Fatal(err)
	}
	repos := projectMappings[0].Repositories

	apiKey := os.Getenv("GITHUB_TOKEN")
	events, err := FetchUserEvents("pbar1", false, apiKey)
	if err != nil {
		log.Fatal(err)
	}

	messagesByDate := make(map[string][]string)
	for _, event := range events {
		if *event.Type != "PushEvent" || !stringInSlice(*event.Repo.Name, repos) {
			continue
		}

		pushEvent, err := ParsePushEvent(event)
		if err != nil {
			log.Fatal(err)
		}

		eventDate := event.CreatedAt.Format("2006-01-02")
		for _, commit := range pushEvent.Commits {
			messagesByDate[eventDate] = append(messagesByDate[eventDate], *commit.Message)
		}
	}

	for date, message := range messagesByDate {
		if _, ok := timeEntryDates[date]; ok {
			continue
		}
		note := strings.Join(message, "\n")
	}
}

// ProjectMapping maps GitHub repositories to Harvest Projects and Tasks
type ProjectMapping struct {
	Project      string
	Task         string
	Repositories []string
}

func readProjectMapping(filename string) ([]*ProjectMapping, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var projectMapping []*ProjectMapping
	err = hcl.Unmarshal(raw, &projectMapping)
	if err != nil {
		return nil, err
	}
	return projectMapping, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func conjectureWorkdays(numDaysBack int) {
	i := 0
	var workdays []time.Time
	for i < numDaysBack {
		t := time.Now().AddDate(0, 0, -i)
		if t.Weekday() != time.Saturday && t.Weekday() != time.Sunday {
			workdays = append(workdays, t)
		}
		i++
	}
	fmt.Println(workdays)
}

// FetchTimeEntries gets all Harvest time entries for the currently authenticated user
func FetchTimeEntries(accountID string, accessToken string) ([]*harvest.TimeEntry, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: accessToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	service := harvest.NewHarvestClient(tc)
	service.AccountId = accountID

	opt := harvest.TimeEntryListOptions{}

	var allEntries []*harvest.TimeEntry
	for {
		timeEntries, resp, err := service.Timesheet.List(ctx, &opt)
		if err != nil {
			log.Fatal(err)
		} else if resp.StatusCode != 200 {
			log.Fatal(resp.Status)
		}

		allEntries = append(allEntries, timeEntries.TimeEntries...)
		if *timeEntries.Page == *timeEntries.TotalPages {
			break
		}
		opt.Page = *timeEntries.NextPage
	}

	return allEntries, nil
}

// FetchUserEvents retries the GitHub events a user has performed
func FetchUserEvents(user string, publicOnly bool, apiKey string) ([]*github.Event, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := github.ListOptions{PerPage: 100}

	var allEvents []*github.Event
	for {
		events, resp, err := client.Activity.ListEventsPerformedByUser(ctx, user, publicOnly, &opt)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allEvents, nil
}

// ParsePushEvent returns only PushEvents from a slice of Events
func ParsePushEvent(event *github.Event) (*github.PushEvent, error) {
	i, err := event.ParsePayload()
	if err != nil {
		return nil, err
	}
	pushEvent, _ := i.(*github.PushEvent)
	return pushEvent, nil
}
