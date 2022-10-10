package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Ltime)

	args := os.Args[1:]

	if len(args) < 1 {
		log.Fatalln("Error. Must specify 1 command line argument")
	}

	resourceToIngest := args[0]

	app, dbConnection := loadApp()
	defer dbConnection.Close()

	switch resourceToIngest {
	case "leagues":
		app.IngestLeagues()
	case "teams":
		app.IngestTeams()
	case "fixtures":
		app.IngestFixtures()
	default:
		log.Fatalln("Unkown cmomand option", resourceToIngest)
	}
}
