package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/wweitzel/top90/internal/clients/apifootball"
)

func (s *Scraper) getEventFromApiFootball(redditPostTitle string, fixtureId int) (apifootball.Event, error) {
	events, err := s.apifbClient.GetEvents(fixtureId)
	if err != nil {
		return apifootball.Event{}, err
	}
	event, err := s.findEvent(redditPostTitle, events)
	if err != nil {
		return apifootball.Event{}, err
	}
	return event, nil
}

func (s *Scraper) findEvent(redditPostTitle string, events []apifootball.Event) (apifootball.Event, error) {
	title := redditPostTitle

	re := regexp.MustCompile(`\([^)]*\)`)
	title = re.ReplaceAllString(title, "")
	title = strings.Replace(title, "great goal", "", -1)
	title = strings.Trim(title, " ")

	parts := strings.Split(title, " ")
	time := parts[len(parts)-1]

	timeParts := strings.Split(time, "+")
	elapsed := timeParts[0]
	elapsed = strings.Replace(elapsed, "'", "", -1)
	elapsed = strings.Replace(elapsed, "’", "", -1)
	elapsed = clean(elapsed)
	var extra string
	if len(timeParts) > 1 {
		extra = timeParts[1]
		extra = strings.Replace(extra, "'", "", -1)
		extra = strings.Replace(extra, "’", "", -1)
		extra = clean(extra)
	}

	elapsedInt, err := strconv.Atoi(elapsed)
	if err != nil {
		return apifootball.Event{}, fmt.Errorf("error converting elapsed time to int: %v", err)
	}
	var extraInt int
	if len(extra) > 0 {
		extraInt, err = strconv.Atoi(extra)
		if err != nil {
			return apifootball.Event{}, fmt.Errorf("error converting extra time to int: %v", err)
		}
	}

	for _, event := range events {
		if (event.Time.Elapsed == elapsedInt && event.Time.Extra == extraInt) ||
			(event.Time.Elapsed == elapsedInt-1 && event.Time.Extra == extraInt) ||
			(event.Time.Elapsed == elapsedInt+1 && event.Time.Extra == extraInt) ||
			(event.Time.Elapsed == elapsedInt && event.Time.Extra == extraInt-1) ||
			(event.Time.Elapsed == elapsedInt && event.Time.Extra == extraInt+1) {
			return event, nil
		}
	}

	return apifootball.Event{}, fmt.Errorf("could not find associated event")
}

func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}
