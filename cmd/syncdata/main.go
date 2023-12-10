package main

import (
	"time"

	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/jsonlogger"
	"github.com/wweitzel/top90/internal/syncdata"
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

	client := init.ApiFootballClient(
		config.ApiFootballRapidApiHost,
		config.ApiFootballRapidApiKey,
		config.ApiFootballRapidApiKeyBackup,
		10*time.Second)

	syncData := syncdata.New(dao, client, logger)

	logger.Info("Syncing leagues...")
	err := syncData.Leagues()
	if err != nil {
		init.Exit("Failed syncing leagues", err)
	}

	logger.Info("Syncing teams...")
	err = syncData.Teams()
	if err != nil {
		init.Exit("Failed syncing teams", err)
	}

	logger.Info("Syncing fixtures...")
	err = syncData.Fixtures()
	if err != nil {
		init.Exit("Failed syncing fixtures", err)
	}
}
