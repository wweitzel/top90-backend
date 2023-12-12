package main

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	db "github.com/wweitzel/top90/internal/db/models"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

var logger *slog.Logger

func main() {
	config := config.Load()
	logger = jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)
	dbConn := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao := init.Dao(dbConn)

	client := init.ApiFootballClient(
		config.ApiFootballRapidApiHost,
		config.ApiFootballRapidApiKey,
		config.ApiFootballRapidApiKeyBackup,
		30*time.Second,
		config.ApiFootballCurrentSeason)

	allGoals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100000}, db.GetGoalsFilter{})
	if err != nil {
		init.Exit("Failed getting goals from database", err)
	}

	for _, goal := range allGoals {
		logger.Info("Processing goal...", "id", goal.Id, "title", goal.RedditPostTitle)

		if int(goal.PlayerId) != 0 {
			logger.Info("Already linked player. Skipping.")
			continue
		}

		events, err := client.GetEvents(int(goal.FixtureId))
		if err != nil {
			init.Exit("Failed getting events from apifootball", err)
		}

		event, err := findEvent(goal, events)
		if err != nil {
			logger.Warn("Could not find event for goal", "error", err)
			time.Sleep(300 * time.Millisecond)
			continue
		}

		logger.Info("Found event", "goalId", goal.Id, "goalTitle", goal.RedditPostTitle, "goalFixtureId", goal.FixtureId, "eventTime", event.Time, "eventPlayer", event.Player)

		player := db.Player{
			Id:   event.Player.ID,
			Name: event.Player.Name,
		}
		_, err = dao.UpsertPlayer(player)
		if err != nil {
			init.Exit("Failed upserting player in db", err)
		}

		goalUpdate := db.Goal{
			Id:         goal.Id,
			Type:       db.NullString(event.Type),
			TypeDetail: db.NullString(event.Detail),
			PlayerId:   db.NullInt(event.Player.ID),
		}
		_, err = dao.UpdateGoal(goal.Id, goalUpdate)
		if err != nil {
			playerApifootball, err := client.GetPlayer(event.Player.ID, 2023)
			if err != nil {
				logger.Error("Failed getting player from apifootball", "error", err)
				continue
			}
			_, err = dao.UpsertPlayer(playerApifootball)
			if err != nil {
				logger.Error("Failed upserting player in database", "error", err)
				continue
			}
			_, err = dao.UpdateGoal(goal.Id, goalUpdate)
			if err != nil {
				logger.Error("Failed updating goal in database", "error", err)
				continue
			}
		}

		time.Sleep(300 * time.Millisecond)
	}
}

func findEvent(goal db.Goal, events []apifootball.Event) (apifootball.Event, error) {
	title := goal.RedditPostTitle
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

	logger.Info("Parsed times", "elapsed", elapsed, "extra", extra)

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
