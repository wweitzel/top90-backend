package main

import (
	"time"

	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	db "github.com/wweitzel/top90/internal/db/models"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

func main() {
	config := config.Load()
	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)

	client := init.ApiFootballClient(
		config.ApiFootballRapidApiHost,
		config.ApiFootballRapidApiKey,
		config.ApiFootballRapidApiKeyBackup,
		10*time.Second)

	dbConn := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao := init.Dao(dbConn)

	query := "SELECT * FROM teams where national = false"
	var teams []db.Team
	err := dbConn.Select(&teams, query)
	if err != nil {
		init.Exit("Failed getting teams from database", err)
	}

	for _, team := range teams {
		logger.Info("Started loading " + team.Name)
		players, err := client.GetPlayers(team.Id, 2023)
		if err != nil {
			logger.Error("Failed getting players from apifootball", "error", err)
		}
		for _, player := range players {
			_, err := dao.UpsertPlayer(player)
			if err != nil {
				logger.Error("Failed upserting players into database", "error", err)
			}
		}
	}
}
