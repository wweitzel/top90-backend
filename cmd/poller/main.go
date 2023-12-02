package main

import (
	"log"

	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	log.Println("Starting Run...")

	config := config.Load()
	goalIngest := scrape.NewGoalIngest(config)
	goalIngest.Run()

	log.Println("Finished.")
}
