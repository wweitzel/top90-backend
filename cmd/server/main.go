package main

import (
	"log"
	"net/http"

	"github.com/wweitzel/top90/internal/server"
)

func main() {
	server.LoadConfig()
	server.InitS3Client()
	server.InitDao()

	r := server.NewRouter()
	http.Handle("/", r)

	port := ":7171"
	log.Println("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, nil)
}
