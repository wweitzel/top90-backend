package main

import (
	"os"
	"strconv"
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

func main() {
	config := config.Load()
	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)
	db := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao := init.Dao(db)

	footballClient := apifootball.NewClient(apifootball.Config{
		Host:    config.ApiFootballRapidApiHost,
		ApiKey:  config.ApiFootballRapidApiKey,
		Timeout: 10 * time.Second,
	})

	leagues, _ := dao.GetLeagues()
	for _, league := range leagues {
		league, _ := footballClient.GetLeague(league.Id)
		leagueUpdate := apifootball.League{CurrentSeason: league.CurrentSeason}
		_, err := dao.UpdateLeague(league.Id, leagueUpdate)
		if err != nil {
			logger.Error("Failed to update league", "error", err)
			os.Exit(1)
		} else {
			logger.Info("Successfully updated league " + league.Name + " current season " + strconv.Itoa(league.CurrentSeason))
		}
	}
}
