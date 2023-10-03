package main

import (
	"log"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/ingest"
)

func main() {
	log.Println("Starting Run...")

	config := top90.LoadConfig()
	goalIngest := ingest.NewGoalIngest(config)
	goalIngest.Run()

	log.Println("Finished.")
}
