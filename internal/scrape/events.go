package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/wweitzel/top90/internal/clients/apifootball"
)

var errNotFound = fmt.Errorf("could not find associated event")

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
	title := cleanTitle(redditPostTitle)
	elapsed, extra, err := timeParts(title)
	if err != nil {
		return apifootball.Event{}, fmt.Errorf("error getting time parts")
	}

	if strings.Contains(title, "yellow card") || strings.Contains(title, "red card") {
		event, err := s.findEventWithDeviation(events, "Card", elapsed, extra, title)
		if err == nil {
			return event, nil
		}
		return apifootball.Event{}, errNotFound
	}

	event, err := s.findEventWithDeviation(events, "Goal", elapsed, extra, title)
	if err == nil {
		return event, nil
	}
	return apifootball.Event{}, errNotFound
}

func (s *Scraper) findEventWithDeviation(events []apifootball.Event, eventType string, elapsedInt int, extraInt int, title string) (apifootball.Event, error) {
	event, err := s.findEventForTime(events, eventType, elapsedInt, extraInt, title)
	if err == nil {
		return event, nil
	}
	if extraInt == 0 {
		event, err = s.findEventForTime(events, eventType, elapsedInt+1, extraInt, title)
		if err == nil {
			return event, nil
		}
		event, err = s.findEventForTime(events, eventType, elapsedInt-1, extraInt, title)
		if err == nil {
			return event, nil
		}
		return apifootball.Event{}, errNotFound
	}
	event, err = s.findEventForTime(events, eventType, elapsedInt, extraInt+1, title)
	if err == nil {
		return event, nil
	}
	event, err = s.findEventForTime(events, eventType, elapsedInt, extraInt-1, title)
	if err == nil {
		return event, nil
	}
	return apifootball.Event{}, errNotFound
}

func (s *Scraper) findEventForTime(events []apifootball.Event, eventType string, elapsedTime int, extraTime int, title string) (apifootball.Event, error) {
	var bestGuessEvent apifootball.Event
	for _, event := range events {
		if event.Time.Elapsed == elapsedTime && event.Time.Extra == extraTime && event.Type == eventType {
			bestGuessEvent = event
			dbPlayer, err := s.dao.GetPlayer(event.Player.ID)
			if err != nil {
				continue
			}
			nameLower := strings.ToLower(dbPlayer.Name)
			firstNameLower := strings.ToLower(dbPlayer.FirstName)
			lastNameLower := strings.ToLower(dbPlayer.LastName)
			// If we can find an event with a player whose name is in the reddit post title, return it immediately
			if (nameLower != "" && strings.Contains(title, nameLower)) ||
				(firstNameLower != "" && strings.Contains(title, firstNameLower)) ||
				(lastNameLower != "" && strings.Contains(title, lastNameLower)) {
				return event, nil
			}
		}
	}
	if bestGuessEvent == (apifootball.Event{}) {
		return apifootball.Event{}, errNotFound
	}
	return bestGuessEvent, nil
}

func timeParts(title string) (elapsed int, extra int, err error) {
	parts := strings.Split(title, " ")
	time := parts[len(parts)-1]
	timeParts := strings.Split(time, "+")

	elapsedStr := timeParts[0]
	elapsedStr = strings.Replace(elapsedStr, "'", "", -1)
	elapsedStr = strings.Replace(elapsedStr, "’", "", -1)
	elapsedStr = clean(elapsedStr)
	var extraStr string
	if len(timeParts) > 1 {
		extraStr = timeParts[1]
		extraStr = strings.Replace(extraStr, "'", "", -1)
		extraStr = strings.Replace(extraStr, "’", "", -1)
		extraStr = clean(extraStr)
	}

	elapsed, err = strconv.Atoi(elapsedStr)
	if err != nil {
		return 0, 0, fmt.Errorf("error converting elapsed time to int: %v", err)
	}
	if len(extraStr) > 0 {
		extra, err = strconv.Atoi(extraStr)
		if err != nil {
			return 0, 0, fmt.Errorf("error converting extra time to int: %v", err)
		}
	}
	return elapsed, extra, nil
}

func cleanTitle(title string) string {
	re := regexp.MustCompile(`\([^)]*\)`)
	title = re.ReplaceAllString(title, "")
	title = strings.Replace(title, "great goal", "", -1)
	title = strings.Trim(title, " ")
	title = strings.ToLower(title)
	return title
}

func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}
